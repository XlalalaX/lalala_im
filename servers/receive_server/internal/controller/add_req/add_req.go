package add_req

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"lalala_im/data/Repo/AddReq"
	"lalala_im/data/Repo/Friend"
	"lalala_im/data/Repo/Group"
	"lalala_im/data/model/MongoModel"
	log "lalala_im/pkg/la_log"
	"lalala_im/pkg/la_rsp"
	"lalala_im/servers/receive_server/internal/db"
	"lalala_im/servers/receive_server/internal/http/middle"
	"time"
)

var (
	addReqRepo      = AddReq.NewIMongoAddReqRepo(db.DB.MongoDB)
	friendRepo      = Friend.NewIFriendRepo(db.DB.MongoDB)
	groupRepo       = Group.NewIGroupRepo(db.DB.MongoDB)
	userToGroupRepo = Group.NewIUserToGroupRepo(db.DB.MongoDB)
)

func GetSelfAddReqList(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	addList, err := addReqRepo.GetListByRecvID(ctx, user.UID)
	if err != nil {
		log.Error("获取添加好友请求列表，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, addList)
	return
}

type getGroupAddReqListReq struct {
	GroupID string `json:"group_id" form:"group_id"`
}

func GetGroupAddReqList(ctx *gin.Context) {
	req := getGroupAddReqListReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	groupAddList, err := addReqRepo.GetListByGroupID(ctx, req.GroupID)
	if err != nil {
		log.Error("获取群组添加好友请求列表，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, groupAddList)
	return
}

type sendAddReqReq struct {
	UID string `json:"uid" form:"uid"`
}

func SendAddReq(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	req := sendAddReqReq{}
	err := ctx.ShouldBind(&req)
	if err != nil || req.UID == "" {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	//判断是否已经是好友
	isFriend, err := friendRepo.GetFriend(ctx, user.UID, req.UID)
	if err != nil {
		log.Error("发送添加好友请求,判断是否是好友，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//如果已经是好友
	if isFriend != nil && !isFriend.Id.IsZero() {
		la_rsp.Success(ctx, nil)
		return
	}
	//判断是否已经发送过添加请求
	addReq, err := addReqRepo.GetListByRecvIDAndSendID(ctx, req.UID, user.UID)
	if err != nil {
		log.Error("发送添加好友请求，判断是否已经发过，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if len(addReq) > 0 {
		la_rsp.Success(ctx, nil)
		return
	}
	//发送添加请求
	err = addReqRepo.Create(ctx, &MongoModel.AddReq{
		SendNickName: user.NickName,
		SendID:       user.UID,
		RecvID:       req.UID,
		GroupID:      "",
	})
	if err != nil {
		log.Error("发送添加好友请求，创建数据库记录失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, nil)
}

type sendGroupAddReqReq struct {
	GroupID string `json:"group_id"`
}

func SendGroupAddReq(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	req := sendGroupAddReqReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	//判断是否已经发送过添加请求
	addReq, err := addReqRepo.GetListByGroupIDAndSendID(ctx, req.GroupID, user.UID)
	if err != nil {
		log.Error("发送添加群组好友请求，判断是否已经发过，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if len(addReq) > 0 {
		la_rsp.Success(ctx, nil)
		return
	}
	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.GroupID)
	if err != nil {
		log.Error("发送添加群组好友请求，获取群组信息失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if groupInfo == nil || groupInfo.Id.IsZero() {
		log.Error("发送添加群组好友请求，群组不存在：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("群组不存在"))
		return
	}
	isMember := false
	for _, v := range groupInfo.UIDList {
		if v == user.UID {
			isMember = true
			break
		}
	}
	if isMember {
		la_rsp.Success(ctx, nil)
		return
	}
	//发送添加请求
	err = addReqRepo.Create(ctx, &MongoModel.AddReq{
		SendNickName: user.NickName,
		SendID:       user.UID,
		RecvID:       "",
		GroupID:      req.GroupID,
		GroupName:    groupInfo.GroupName,
	})
	if err != nil {
		log.Error("发送添加群组好友请求，创建数据库记录失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, nil)
}

type handleAddReqReq struct {
	AddReqID string `json:"add_req_id"`
	IsAgree  bool   `json:"is_agree"`
}

func HandleAddReq(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	req := handleAddReqReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	//获取添加请求
	addReq, err := addReqRepo.GetInfoByID(ctx, req.AddReqID)
	if err != nil {
		log.Error("处理添加好友or群组请求，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if addReq == nil || addReq.Id.IsZero() {
		log.Error("处理添加好友or群组请求，添加请求不存在", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("添加请求不存在"))
		return
	}
	//如果是添加好友请求
	if addReq.GroupID == "" {
		//判断是否已经是好友
		isFriend, err := friendRepo.GetFriend(ctx, user.UID, addReq.SendID)
		if err != nil {
			log.Error("处理添加好友请求，判断是否是好友，查询数据库失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		//如果已经是好友
		if isFriend != nil && !isFriend.Id.IsZero() {
			addList, err := addReqRepo.GetListByRecvIDAndSendID(ctx, user.UID, addReq.SendID)
			if err != nil {
				log.Error("处理添加好友请求，获取需要删除的请求记录失败：", err)
				la_rsp.Failed(ctx, la_rsp.ERROR, err)
				return
			}
			for _, v := range addList {
				err = addReqRepo.Delete(ctx, v.Id.Hex())
				if err != nil {
					log.Error("处理添加好友请求，删除请求记录失败：", err)
					la_rsp.Failed(ctx, la_rsp.ERROR, err)
					return
				}
			}
			la_rsp.Success(ctx, nil)
			return
		}
		//如果不是好友
		if !req.IsAgree {
			//不同意添加,删除请求记录
			addList, err := addReqRepo.GetListByRecvIDAndSendID(ctx, user.UID, addReq.SendID)
			if err != nil {
				log.Error("处理添加好友请求，获取需要删除的请求记录失败：", err)
				la_rsp.Failed(ctx, la_rsp.ERROR, err)
				return
			}
			for _, v := range addList {
				err = addReqRepo.Delete(ctx, v.Id.Hex())
				if err != nil {
					log.Error("处理添加好友请求，删除请求记录失败：", err)
					la_rsp.Failed(ctx, la_rsp.ERROR, err)
					return
				}
			}
			la_rsp.Success(ctx, nil)
		}
		//同意添加
		err = friendRepo.CreateFriend(ctx, user.UID, addReq.SendID)
		if err != nil {
			log.Error("处理添加好友请求，创建好友关系失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		//需要刷新缓存，直接删除了懒得搞
		err = db.DB.Rdb.SAdd(ctx, "friend_list:"+user.UID, addReq.SendID).Err()
		if err != nil && err != redis.Nil {
			log.Error("添加好友请求，更新缓存数据失败：", err)
		}
		err = db.DB.Rdb.Del(ctx, "friend_list:"+addReq.SendID, user.UID).Err()
		if err != nil && err != redis.Nil {
			log.Error("添加好友请求，更新缓存数据失败：", err)
		}
		//删除请求记录
		addList, err := addReqRepo.GetListByRecvIDAndSendID(ctx, user.UID, addReq.SendID)
		if err != nil {
			log.Error("处理添加好友请求，获取需要删除的请求记录失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		for _, v := range addList {
			err = addReqRepo.Delete(ctx, v.Id.Hex())
			if err != nil {
				log.Error("处理添加好友请求，删除请求记录失败：", err)
				la_rsp.Failed(ctx, la_rsp.ERROR, err)
				return
			}
		}
		la_rsp.Success(ctx, nil)
		return
	}
	//如果是添加群组请求
	//判断是否已经是群组成员
	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, addReq.GroupID)
	if err != nil {
		log.Error("处理添加群组请求，判断是否是群组成员，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if groupInfo == nil || groupInfo.Id.IsZero() {
		log.Error("处理添加群组请求，群组不存在", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("群组不存在"))
		return
	}
	isAdmin := false
	for _, v := range groupInfo.Admin {
		if v == user.UID {
			isAdmin = true
		}
	}
	if !isAdmin {
		log.Error("处理添加群组请求，不是群组管理员", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("不是群组管理员"))
		return
	}
	isMember := false
	for _, v := range groupInfo.UIDList {
		if v == addReq.SendID {
			isMember = true
			break
		}
	}
	//如果已经是群组成员或者不同意添加
	if isMember || !req.IsAgree {
		//删除请求记录
		senderGroupList, err := addReqRepo.GetListByGroupIDAndSendID(ctx, addReq.GroupID, addReq.SendID)
		if err != nil {
			log.Error("处理添加群组请求，获取需要删除的请求记录失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		for _, v := range senderGroupList {
			err = addReqRepo.Delete(ctx, v.Id.Hex())
			if err != nil {
				log.Error("处理添加群组请求，删除请求记录失败：", err)
				la_rsp.Failed(ctx, la_rsp.ERROR, err)
				return
			}
		}
		la_rsp.Success(ctx, nil)
	}
	//如果不是群组成员
	err = userToGroupRepo.CreateUserToGroup(ctx, &MongoModel.UserToGroup{
		GroupID:  addReq.GroupID,
		UID:      addReq.SendID,
		Nickname: addReq.SendNickName,
		JoinTime: time.Now(),
	})
	if err != nil {
		log.Error("处理添加群组请求，创建用户-群记录失败,添加群组成员失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	groupInfo.UIDList = append(groupInfo.UIDList, addReq.SendID)
	err = groupRepo.UpdateGroup(ctx, groupInfo)
	if err != nil {
		log.Error("处理添加群组请求，更新群组信息失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	err = db.DB.Rdb.SAdd(ctx, "group_members:"+groupInfo.GroupID, addReq.SendID).Err()
	if err != nil {
		log.Error("处理添加群组请求，更新redis中群成员列表失败：", err)
		return
	}
	//删除请求记录
	senderGroupList, err := addReqRepo.GetListByGroupIDAndSendID(ctx, addReq.GroupID, addReq.SendID)
	if err != nil {
		log.Error("处理添加群组请求，获取需要删除的请求记录失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	for _, v := range senderGroupList {
		err = addReqRepo.Delete(ctx, v.Id.Hex())
		if err != nil {
			log.Error("处理添加群组请求，删除请求记录失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}
	la_rsp.Success(ctx, nil)
}
