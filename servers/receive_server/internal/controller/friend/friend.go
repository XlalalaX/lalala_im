package friend

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/qiniu/qmgo"
	"lalala_im/data/Repo/Friend"
	log "lalala_im/pkg/la_log"
	"lalala_im/pkg/la_rsp"
	"lalala_im/servers/receive_server/internal/db"
	"lalala_im/servers/receive_server/internal/http/middle"
)

var (
	//userRepo   = User.NewIUserRepo(db.DB.MongoDB)
	friendRepo = Friend.NewIFriendRepo(db.DB.MongoDB)
)

func GetFriendList(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	//从缓存获取
	friends, err := db.DB.Rdb.SMembers(ctx, "friend_list:"+user.UID).Result()
	if err != nil && err != redis.Nil {
		log.Error("获取好友列表请求，查询redis失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if len(friends) != 0 {
		la_rsp.Success(ctx, map[string]interface{}{
			"friend_list": friends,
		})
		return
	}

	friendsList, err := friendRepo.GetFriendListByUID(ctx, user.UID)
	if err != nil {
		log.Error("获取好友列表请求，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	for i := 0; i < len(friendsList); i++ {
		if friendsList[i].FirstUID != user.UID {
			friends = append(friends, friendsList[i].FirstUID)
		} else {
			friends = append(friends, friendsList[i].SecondUID)
		}
	}
	if len(friends) == 0 {
		la_rsp.Success(ctx, map[string]interface{}{
			"friend_list": friends,
		})
		return
	}
	//数据库查询后写入缓存，不影响返回
	err = db.DB.Rdb.SAdd(ctx, "friend_list:"+user.UID, friends).Err()
	if err != nil {
		log.Error("获取好友列表请求，更新redis缓存失败：", err)
	}
	la_rsp.Success(ctx, friends)
}

type friendReq struct {
	FriendId string `json:"friend_id" form:"friend_id"`
}

func IsFriend(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	req := &friendReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		log.Error("判断是否为好友请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	is, err := db.DB.Rdb.SIsMember(ctx, "friend_list:"+user.UID, req.FriendId).Result()
	if err != nil && err != redis.Nil {
		log.Error("判断是否为好友请求，查询redis失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//有缓存数据直接返回
	if err == nil {
		la_rsp.Success(ctx, map[string]interface{}{
			"is": is,
		})
		return
	}

	friendsList, err := friendRepo.GetFriendListByUID(ctx, user.UID)
	if err != nil {
		log.Error("获取好友列表请求，查询数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	friends := make([]string, 0, len(friendsList))
	for i := 0; i < len(friendsList); i++ {
		if friendsList[i].FirstUID != user.UID {
			if friendsList[i].FirstUID == req.FriendId {
				is = true
			}
			friends = append(friends, friendsList[i].FirstUID)
		} else {
			if friendsList[i].SecondUID == req.FriendId {
				is = true
			}
			friends = append(friends, friendsList[i].SecondUID)
		}
	}
	//数据库查询后写入缓存，不影响返回
	err = db.DB.Rdb.SAdd(ctx, "friend_list:"+user.UID, friends).Err()
	if err != nil {
		log.Error("获取好友列表请求，更新redis缓存失败：", err)
	}
	la_rsp.Success(ctx, map[string]interface{}{
		"is": is,
	})
}

func AddFriend(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	req := &friendReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		log.Error("添加好友请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	friend, err := friendRepo.GetFriend(ctx, req.FriendId, user.UID)
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		log.Error("添加好友请求，从数据库查询是否为好友错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	//已经为好友，直接返回
	if friend != nil && !friend.Id.IsZero() {
		la_rsp.Success(ctx, nil)
		return
	}
	err = friendRepo.CreateFriend(ctx, req.FriendId, user.UID)
	if err != nil {
		log.Error("添加好友请求，添加好友到数据库错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	//需要刷新缓存，直接删除了懒得搞
	err = db.DB.Rdb.Del(ctx, "friend_list:"+user.UID).Err()
	if err != nil && err != redis.Nil {
		log.Error("添加好友请求，清除缓存数据失败：", err)
	}
	err = db.DB.Rdb.Del(ctx, "friend_list:"+req.FriendId).Err()
	if err != nil && err != redis.Nil {
		log.Error("添加好友请求，清除缓存数据失败：", err)
	}
	la_rsp.Success(ctx, nil)
	return
}

func DelFriend(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	req := &friendReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		log.Error("添加好友请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	friend, err := friendRepo.GetFriend(ctx, user.UID, req.FriendId)
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		log.Error("添加好友请求，从数据库查询好友信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if friend.Id.IsZero() {
		la_rsp.Success(ctx, nil)
		return
	}
	err = friendRepo.DelFriend(ctx, friend.Id.Hex())
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		log.Error("添加好友请求，从数据库删除好友链接错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	err = db.DB.Rdb.Del(ctx, "friend_list:"+user.UID).Err()
	if err != nil {
		log.Error("添加好友请求，移除redis缓存失败：", err)
	}
	err = db.DB.Rdb.Del(ctx, "friend_list:"+req.FriendId).Err()
	if err != nil {
		log.Error("添加好友请求，移除redis缓存失败：", err)
	}
	la_rsp.Success(ctx, nil)
}
