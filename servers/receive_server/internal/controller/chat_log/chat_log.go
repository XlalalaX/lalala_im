package chat_log

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"lalala_im/data/Repo/ChatLog"
	"lalala_im/data/Repo/Friend"
	"lalala_im/data/Repo/SaveLog"
	log "lalala_im/pkg/la_log"
	"lalala_im/pkg/la_rsp"
	"lalala_im/proto/pb_msg"
	"lalala_im/servers/receive_server/internal/db"
	"lalala_im/servers/receive_server/internal/http/middle"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	chatRepo   = ChatLog.NewIChatLogRepo(db.DB.MongoDB)
	saveRepo   = SaveLog.NewISaveLogRepo(db.DB.MongoDB)
	friendRepo = Friend.NewIFriendRepo(db.DB.MongoDB)
)

type getRecentChatLogListReq struct {
	ChatId   string `json:"chat_id" form:"chat_id"`
	ChatType int    `json:"chat_type" form:"chat_type"` //0单人，1群聊
}

func GetRecentChatLogList(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &getRecentChatLogListReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		log.Error(fmt.Sprintf("获取临时聊天记录,参数解析错误：%+v", err))
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	//校验一遍是不是有权限（是不是自己的聊天记录）
	if req.ChatType == 0 {
		ids := strings.Split(req.ChatId, "_")
		if len(ids) != 2 {
			log.Error(fmt.Sprintf("获取临时聊天记录,参数解析错误,非法群聊id：%+v", err))
			la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
			return
		}
		friendId := ids[0]
		if ids[0] == user.UID {
			friendId = ids[1]
		}
		isFriend := false
		//判断发起请求的用户是不是请求用户的好友，不是直接返回（不能随便查别人聊天记录）
		friendsList, err := friendRepo.GetFriendListByUID(ctx, friendId)
		if err != nil {
			log.Error("获取好友列表请求，查询数据库失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		friends := make([]string, 0, len(friendsList))
		for i := 0; i < len(friendsList); i++ {
			if friendsList[i].FirstUID == user.UID || friendsList[i].SecondUID == user.UID {
				isFriend = true
			}
			if friendsList[i].FirstUID == friendId {
				friends = append(friends, friendsList[i].SecondUID)
			} else {
				friends = append(friends, friendsList[i].FirstUID)
			}
		}
		if !isFriend {
			la_rsp.Failed(ctx, la_rsp.AuthError, errors.New("没权限获取不是好友的聊天记录"))
			return
		}

		//数据库查询后写入缓存，不影响返回
		err = db.DB.Rdb.SAdd(ctx, "friend_list:"+user.UID, friendId).Err()
		if err != nil {
			log.Error("获取好友列表请求，更新redis缓存失败：", err)
		}
		//数据库查询后写入缓存，不影响返回
		err = db.DB.Rdb.SAdd(ctx, "friend_list:"+friendId, friends).Err()
		if err != nil {
			log.Error("获取好友列表请求，更新redis缓存失败：", err)
		}
	} else {
		ismember, err := db.DB.Rdb.SIsMember(ctx, "group_members:"+req.ChatId, user.UID).Result()
		if err != nil && err != redis.Nil {
			log.Error("获取临时聊天记录，判断是否为群员，查询redis错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		if !ismember {
			la_rsp.Failed(ctx, la_rsp.AuthError, errors.New("没权限获取不在群组的聊天记录"))
			return
		}
	}

	tempChatLog, err := chatRepo.GetChatLogByChatId(ctx, req.ChatId)
	if err != nil {
		log.Error(fmt.Sprintf("获取临时聊天记录，查询数据库失败：%+v", err))
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	for i := 0; i < len(tempChatLog.Msg); i++ {
		msg := &pb_msg.Msg{}
		err := proto.Unmarshal(*tempChatLog.Msg[i].Msg, msg)
		if err != nil {
			log.Error(fmt.Sprintf("解析错误：%+v", err))
		}
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error(fmt.Sprintf("推送历史消息出现panic%+v", err))
				return
			}
		}()
		addr, err := db.DB.Rdb.Get(ctx, "user_conn:"+user.UID).Result()
		if err != nil || addr == "" {
			log.Error(fmt.Sprintf("推送历史消息出错,用户不在线，错误：%+v", err))
			return
		}
		conn, err := grpc.Dial(addr, grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*100), // 最大消息接收大小为100MB
			grpc.MaxCallSendMsgSize(1024*1024*100), // 最大消息发送大小为100MB
		), grpc.WithInsecure(),
			grpc.WithMaxMsgSize(1024*1024*100), //最大消息大小为100MB
		)
		if err != nil {
			log.Error(fmt.Sprintf("创建grpcClient失败，addr：%s,错误：%+v", addr, err))
			return
		}
		cli := pb_msg.NewMsgServerClient(conn)
		for i := 0; i < len(tempChatLog.Msg); i++ {
			//解析
			msg := &pb_msg.Msg{}
			err := proto.Unmarshal(*tempChatLog.Msg[i].Msg, msg)
			if err != nil {
				log.Error(fmt.Sprintf("解析错误：%+v", err))
			}
			msg.SelfID = user.UID
			//推送
			rsp, err := cli.Push(ctx, msg)
			//返回err不为nil或者消息解析错误或者用户不存在，直接跳出
			if err != nil || rsp.ErrCode == -1 || rsp.ErrCode == -2 {
				log.Error(fmt.Sprintf("消费消息出错,push错误，消息：%+v，响应：%+v，错误：%+v", msg, rsp, err))
				continue
			}
		}
	}()
	la_rsp.Success(ctx, nil)
}

type getSaveLogListReq struct {
	ChatId    string `json:"chat_id" form:"chat_id"`
	ChatType  int    `json:"chat_type" form:"chat_type"` //0单人，1群聊
	StartTime int64  `json:"start_time" form:"start_time"`
	EndTime   int64  `json:"end_time" form:"end_time"`
	Limit     int64  `json:"limit" form:"limit"`
}

func GetSaveLogList(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &getSaveLogListReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		log.Error(fmt.Sprintf("获取指定时间区间聊天记录,参数解析错误：%+v", err))
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	//校验一遍是不是有权限（是不是自己的聊天记录）
	if req.ChatType == 0 {
		ids := strings.Split(req.ChatId, "_")
		if len(ids) != 2 {
			log.Error(fmt.Sprintf("获取指定时间区间聊天记录,参数解析错误,非法群聊id：%+v", err))
			la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
			return
		}
		friendId := ids[0]
		if ids[0] == user.UID {
			friendId = ids[1]
		}
		//判断发起请求的用户是不是请求用户的好友，不是直接返回（不能随便查别人聊天记录）
		is, err := db.DB.Rdb.SIsMember(ctx, "friend_list:"+friendId, user.UID).Result()
		if err != nil && err != redis.Nil {
			log.Error("获取指定时间区间聊天记录，判断是否为好友，查询redis失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		if !is {
			la_rsp.Failed(ctx, la_rsp.AuthError, errors.New("没权限获取他人聊天记录"))
			return
		}
	} else {
		ismember, err := db.DB.Rdb.SIsMember(ctx, "group_members:"+req.ChatId, user.UID).Result()
		if err != nil && err != redis.Nil {
			log.Error("获取指定时间区间聊天记录，判断是否为群员，查询redis错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		if !ismember {
			la_rsp.Failed(ctx, la_rsp.AuthError, errors.New("没权限获取不在群组的聊天记录"))
			return
		}
	}

	tempLogList, err := saveRepo.GetSaveLogList(ctx, req.ChatId, req.StartTime, req.EndTime, req.Limit)
	if err != nil {
		log.Error("获取指定时间区间聊天记录，查询数据库错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, tempLogList)
	return
}

// UploadChatFile 上传聊天文件并返回实际存储路径
func UploadChatFile(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	form, err := ctx.MultipartForm()
	if err != nil {
		log.Error("上传文件，解析form错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if len(form.File["file"]) == 0 || len(form.Value["file_size"]) == 0 {
		log.Error("上传文件，解析form提取文件错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	f := form.File["file"][0]
	fileSize := form.Value["file_size"][0]
	if fileSize != strconv.Itoa(int(f.Size)) {
		log.Error("上传文件，文件大小一致错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	isExist, err := db.DB.AliyunOss.IsObjectExist("user/" + user.UID + "/chat_file/" + f.Filename + filepath.Ext(f.Filename))
	if err != nil {
		log.Error("上传文件，判断文件是否已经存在错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	startNum := 0
	//要是文件已经存在，添加备注+1
	for isExist {
		startNum++
		isExist, err = db.DB.AliyunOss.IsObjectExist("user/" + user.UID + "/chat_file/" + f.Filename + "(" + strconv.Itoa(startNum) + ")" + filepath.Ext(f.Filename))
		if err != nil {
			log.Error("上传文件，判断文件是否已经存在错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}
	fReader, err := f.Open()
	if err != nil {
		log.Error("上传文件到阿里云时，打开文件错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if startNum != 0 {
		f.Filename = f.Filename + "(" + strconv.Itoa(startNum) + ")" + filepath.Ext(f.Filename)
	}
	//获取url
	err = db.DB.AliyunOss.PutObject("user/"+user.UID+"/chat_file/"+f.Filename, fReader)
	if err != nil {
		log.Error("上传文件到阿里云时，上传文件错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, f.Filename)
}

type getChatFileUrlReq struct {
	FileName string `form:"file_name" json:"file_name" binding:"required"`
	UID      string `form:"uid" json:"uid" binding:"required"`
}

func GetChatFileUrl(ctx *gin.Context) {
	req := getChatFileUrlReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		log.Error("获取聊天文件url，参数解析错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	//获取url
	url, err := db.DB.AliyunOss.SignURL("user/"+req.UID+"/chat_file/"+req.FileName, oss.HTTPGet, 60*60*24*7)
	if err != nil {
		log.Error("获取聊天文件url，获取url错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, url)
}
