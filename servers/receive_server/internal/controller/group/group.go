package group

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qiniu/qmgo"
	"lalala_im/data/Repo/Group"
	"lalala_im/data/model/MongoModel"
	log "lalala_im/pkg/la_log"
	"lalala_im/pkg/la_rsp"
	"lalala_im/servers/receive_server/internal/db"
	"lalala_im/servers/receive_server/internal/http/middle"
	"time"
)

var (
	groupRepo       = Group.NewIGroupRepo(db.DB.MongoDB)
	userToGroupRepo = Group.NewIUserToGroupRepo(db.DB.MongoDB)
)

func GetSelfGroupList(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	grouplist, err := userToGroupRepo.GetUserToGroupListByUID(ctx, user.UID)
	if err != nil {
		log.Error(fmt.Sprintf("获取用户所加群列表，从数据库查询失败：%+v", err))
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	m := map[string]bool{}
	for i := 0; i < len(grouplist); i++ {
		m[grouplist[i].GroupID] = true
	}
	groupIdList := []string{}
	for k, v := range m {
		if v {
			groupIdList = append(groupIdList, k)
		}
	}
	la_rsp.Success(ctx, map[string]interface{}{
		"group_list": groupIdList,
	})
	return
}

type getGroupReq struct {
	GroupId string `json:"group_id" form:"group_id"`
}

// GetGroupInfo 获取群详情
func GetGroupInfo(ctx *gin.Context) {
	req := &getGroupReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	groupInfo := &MongoModel.Group{}
	//从缓存获取
	err = db.DB.GetObjectFromRedis(ctx, "group_info:"+req.GroupId, groupInfo)
	if err != nil && err != redis.Nil {
		log.Error("获取群信息请求，查询redis失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if groupInfo.Id.IsZero() {
		//缓存没有，刷新缓存
		groupInfo, err = reFlashGroupInfo(ctx, req.GroupId)
		if err != nil {
			log.Error("获取群信息请求，查询数据库失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		if groupInfo == nil || groupInfo.Id.IsZero() || groupInfo.State != 0 {
			la_rsp.Success(ctx, nil)
			return
		}
	}

	la_rsp.Success(ctx, groupInfo)
}

type creatGroupReq struct {
	GroupName string `json:"group_name"`
}

// CreatGroup 新建群
func CreatGroup(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &creatGroupReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	groupId := ""
	for {
		groupId = uuid.NewString()
		group, err := groupRepo.GetGroupInfoByGroupId(ctx, groupId)
		if err != nil && err != qmgo.ErrNoSuchDocuments {
			log.Error("创建群请求，查询数据库失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		//群不存在跳出
		if group == nil || group.Id.IsZero() {
			break
		}
	}
	err = userToGroupRepo.CreateUserToGroup(ctx, &MongoModel.UserToGroup{
		GroupID:  groupId,
		UID:      user.UID,
		Nickname: user.NickName,
		JoinTime: time.Now(),
	})
	if err != nil {
		log.Error("创建群请求，创建用户-群记录到数据库失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	groupInfo := &MongoModel.Group{
		GroupID:   groupId,
		GroupName: req.GroupName,
		UIDList:   []string{user.UID},
		Admin:     []string{user.UID},
		BanUID:    nil,
		OwnerUID:  user.UID,
		MaxMember: 100,
		State:     0,
	}
	err = groupRepo.CreateGroup(ctx, groupInfo)
	if err != nil {
		if err != nil && err != qmgo.ErrNoSuchDocuments {
			log.Error("创建群请求，数据库创建新群失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}

	la_rsp.Success(ctx, groupInfo)
}

type addMemberReq struct {
	GroupId   string `json:"group_id"`
	MemberUID string `json:"member_uid"`
	NickName  string `json:"nick_name"`
}

// AddMember 群里加新人
func AddMember(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &addMemberReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	//看缓存还在不在
	nEx, err := db.DB.Rdb.Exists(ctx, []string{"group_admin:" + req.GroupId, "group_members:" + req.GroupId}...).Result()
	if err != nil {
		log.Error("群加人请求，从redis判断缓存是否存在失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if nEx != 2 {
		//其实已经拿到了群信息了但是懒得写
		_, err = reFlashGroupInfo(ctx, req.GroupId)
		if err != nil {
			log.Error("群加人请求，刷新缓存失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}

	is, err := db.DB.Rdb.SIsMember(ctx, "group_admin:"+req.GroupId, user.UID).Result()
	if err != nil {
		log.Error("群加人请求，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if !is {
		la_rsp.Failed(ctx, la_rsp.AuthError, nil)
		return
	}
	ismember, err := db.DB.Rdb.SIsMember(ctx, "group_members:"+req.GroupId, req.MemberUID).Result()
	if err != nil {
		log.Error("群加人请求，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//已经是群成员，直接返回
	if ismember {
		la_rsp.Success(ctx, nil)
		return
	}

	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.GroupId)
	//err不为空或者没找到都返回错误
	if err != nil {
		log.Error("群加人请求，从数据库查询群信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if groupInfo == nil || groupInfo.Id.IsZero() || groupInfo.State != 0 {
		la_rsp.Failed(ctx, la_rsp.GroupNotExist, err)
		return
	}
	//先获取用户与群的记录,获取到则视为用户已经在群内，直接成功
	userAndGroup, err := userToGroupRepo.GetUserToGroupInfoByGroupIdAndUID(ctx, groupInfo.GroupID, req.MemberUID)
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if !userAndGroup.Id.IsZero() {
		la_rsp.Success(ctx, nil)
		return
	}
	//先创建用户与群的记录
	err = userToGroupRepo.CreateUserToGroup(ctx, &MongoModel.UserToGroup{
		GroupID:  groupInfo.GroupID,
		UID:      req.MemberUID,
		Nickname: req.NickName,
		JoinTime: time.Now(),
	})
	if err != nil {
		log.Error("群加人请求，创建用户与群组的数据库信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	groupInfo.UIDList = append(groupInfo.UIDList, req.MemberUID)
	err = groupRepo.UpdateGroup(ctx, groupInfo)
	if err != nil {
		log.Error("群加人请求，更新数据库群信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//刷新缓存
	go reFlashGroupInfo(ctx, req.GroupId)
	la_rsp.Success(ctx, nil)
}

type delMemberReq struct {
	GroupId   string `json:"group_id"`
	MemberUID string `json:"member_uid"`
}

// DelMember 管理员删人
func DelMember(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &delMemberReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	//看缓存还在不在
	nEx, err := db.DB.Rdb.Exists(ctx, []string{"group_admin:" + req.GroupId, "group_members:" + req.GroupId}...).Result()
	if err != nil {
		log.Error("群删人请求，从redis判断缓存是否存在失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if nEx != 2 {
		//其实已经拿到了群信息了但是懒得写
		_, err = reFlashGroupInfo(ctx, req.GroupId)
		if err != nil {
			log.Error("群删人请求，刷新缓存失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}

	is, err := db.DB.Rdb.SIsMember(ctx, "group_admin:"+req.GroupId, user.UID).Result()
	if err != nil {
		log.Error("群删人请求，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if !is {
		la_rsp.Failed(ctx, la_rsp.AuthError, nil)
		return
	}
	//判断是否被删的也为管理员，是的话不给删
	is, err = db.DB.Rdb.SIsMember(ctx, "group_admin:"+req.GroupId, req.MemberUID).Result()
	if err != nil {
		log.Error("群删人请求，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if is {
		la_rsp.Failed(ctx, la_rsp.AuthError, nil)
		return
	}

	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.GroupId)
	if err != nil {
		log.Error("群删人请求，从数据库获取群信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//群不存在返回错误
	if groupInfo == nil || groupInfo.Id.IsZero() || groupInfo.State != 0 {
		la_rsp.Failed(ctx, la_rsp.GroupNotExist, nil)
		return
	}

	//先删除用户与群的记录
	err = userToGroupRepo.DelUserToGroupByUIDAndGroupID(ctx, req.MemberUID, req.GroupId)
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		log.Error("群删人请求，删除用户与群组的数据库记录错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	newMemberList := []string{}
	newBanList := []string{}
	for i := 0; i < len(groupInfo.UIDList); i++ {
		if groupInfo.UIDList[i] != req.MemberUID {
			newMemberList = append(newMemberList, groupInfo.UIDList[i])
		}
	}
	for i := 0; i < len(groupInfo.BanUID); i++ {
		if groupInfo.BanUID[i] != req.MemberUID {
			newBanList = append(newBanList, groupInfo.BanUID[i])
		}
	}
	groupInfo.UIDList = newMemberList
	groupInfo.BanUID = newBanList

	err = groupRepo.UpdateGroup(ctx, groupInfo)
	if err != nil {
		log.Error("群删人请求，更新数据库记录错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	go reFlashGroupInfo(ctx, req.GroupId)
	la_rsp.Success(ctx, nil)
}

type removeSelfReq struct {
	GroupId string `json:"group_id"`
}

// RemoveSelf 自己退群
func RemoveSelf(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &removeSelfReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	//看缓存还在不在
	nEx, err := db.DB.Rdb.Exists(ctx, []string{"group_members:" + req.GroupId}...).Result()
	if err != nil {
		log.Error("自己退群请求，从redis判断缓存是否存在失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if nEx != 1 {
		//其实已经拿到了群信息了但是懒得写
		_, err = reFlashGroupInfo(ctx, req.GroupId)
		if err != nil {
			log.Error("自己退群请求，刷新缓存失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}

	is, err := db.DB.Rdb.SIsMember(ctx, "group_members:"+req.GroupId, user.UID).Result()
	if err != nil {
		log.Error("自己退群请求，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//都不在群里退个鸡儿
	if !is {
		la_rsp.Success(ctx, nil)
		return
	}

	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.GroupId)
	if err != nil {
		log.Error("自己退群请求，从数据库获取群信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//群不存在返回错误
	if groupInfo == nil || groupInfo.Id.IsZero() || groupInfo.State != 0 {
		la_rsp.Failed(ctx, la_rsp.GroupNotExist, nil)
		return
	}
	//群主，解散群
	if user.UID == groupInfo.OwnerUID {
		groupInfo.State = 1
	}

	//
	newMemberList := []string{}
	newAdminList := []string{}
	newBanList := []string{}
	for i := 0; i < len(groupInfo.UIDList); i++ {
		if groupInfo.UIDList[i] != user.UID {
			newMemberList = append(newMemberList, groupInfo.UIDList[i])
		}
	}
	for i := 0; i < len(groupInfo.Admin); i++ {
		if groupInfo.Admin[i] != user.UID {
			newAdminList = append(newAdminList, groupInfo.Admin[i])
		}
	}
	for i := 0; i < len(groupInfo.BanUID); i++ {
		if groupInfo.BanUID[i] != user.UID {
			newBanList = append(newBanList, groupInfo.BanUID[i])
		}
	}
	groupInfo.UIDList = newMemberList
	groupInfo.Admin = newAdminList
	groupInfo.BanUID = newBanList

	err = userToGroupRepo.DelUserToGroupByUIDAndGroupID(ctx, user.UID, req.GroupId)
	if err != nil {
		log.Error("自己退群请求，删除用户与群组的数据库记录错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	err = groupRepo.UpdateGroup(ctx, groupInfo)
	if err != nil {
		log.Error("自己退群请求，更新数据库记录错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	go reFlashGroupInfo(ctx, req.GroupId)
	la_rsp.Success(ctx, nil)
}

type addBanReq struct {
	GroupId   string `json:"group_id"`
	MemberUID string `json:"member_uid"`
}

// AddBan 群管理ban人
func AddBan(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &addBanReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	//看缓存还在不在
	nEx, err := db.DB.Rdb.Exists(ctx, []string{"group_admin:" + req.GroupId, "group_members:" + req.GroupId}...).Result()
	if err != nil {
		log.Error("群管理ban人请求，从redis判断缓存是否存在失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if nEx != 2 {
		//其实已经拿到了群信息了但是懒得写
		_, err = reFlashGroupInfo(ctx, req.GroupId)
		if err != nil {
			log.Error("群管理ban人请求，刷新缓存失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}

	is, err := db.DB.Rdb.SIsMember(ctx, "group_admin:"+req.GroupId, user.UID).Result()
	if err != nil {
		log.Error("群管理ban人请求，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if !is {
		la_rsp.Failed(ctx, la_rsp.AuthError, nil)
		return
	}
	//判断是否被ban的也为管理员，是的话不给ban
	is, err = db.DB.Rdb.SIsMember(ctx, "group_admin:"+req.GroupId, req.MemberUID).Result()
	if err != nil {
		log.Error("群管理ban人请求，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if is {
		la_rsp.Failed(ctx, la_rsp.AuthError, nil)
		return
	}

	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.GroupId)
	if err != nil {
		log.Error("群管理ban人请求，从数据库获取群信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//群不存在返回错误
	if groupInfo == nil || groupInfo.Id.IsZero() || groupInfo.State != 0 {
		la_rsp.Failed(ctx, la_rsp.GroupNotExist, nil)
		return
	}

	isbaned := false
	for i := 0; i < len(groupInfo.BanUID); i++ {
		if groupInfo.BanUID[i] == req.MemberUID {
			isbaned = true
		}
	}
	if !isbaned {
		groupInfo.BanUID = append(groupInfo.BanUID, req.MemberUID)
	}

	err = groupRepo.UpdateGroup(ctx, groupInfo)
	if err != nil {
		log.Error("群管理ban人请求，更新数据库记录错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	go reFlashGroupInfo(ctx, req.GroupId)
	la_rsp.Success(ctx, nil)
}

type delBanReq struct {
	GroupId   string `json:"group_id"`
	MemberUID string `json:"member_uid"`
}

// DelBan 群管理解封ban人
func DelBan(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &delBanReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	//看缓存还在不在
	nEx, err := db.DB.Rdb.Exists(ctx, []string{"group_admin:" + req.GroupId, "group_members:" + req.GroupId}...).Result()
	if err != nil {
		log.Error("群管理解封ban人，从redis判断缓存是否存在失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if nEx != 2 {
		//其实已经拿到了群信息了但是懒得写
		_, err = reFlashGroupInfo(ctx, req.GroupId)
		if err != nil {
			log.Error("群管理解封ban人，刷新缓存失败：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}

	is, err := db.DB.Rdb.SIsMember(ctx, "group_admin:"+req.GroupId, user.UID).Result()
	if err != nil {
		log.Error("群管理解封ban人，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if !is {
		la_rsp.Failed(ctx, la_rsp.AuthError, nil)
		return
	}
	//判断被解ban的是否在被banlist里
	is, err = db.DB.Rdb.SIsMember(ctx, "group_bans:"+req.GroupId, req.MemberUID).Result()
	if err != nil {
		log.Error("群管理解封ban人，从redis判断是否为管理员错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//不在直接返回
	if !is {
		la_rsp.Success(ctx, nil)
		return
	}

	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.GroupId)
	if err != nil {
		log.Error("群管理解封ban人，从数据库获取群信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//群不存在返回错误
	if groupInfo == nil || groupInfo.Id.IsZero() || groupInfo.State != 0 {
		la_rsp.Failed(ctx, la_rsp.GroupNotExist, nil)
		return
	}

	newBanList := []string{}
	for i := 0; i < len(groupInfo.BanUID); i++ {
		if groupInfo.BanUID[i] != req.MemberUID {
			newBanList = append(newBanList, groupInfo.BanUID[i])
		}
	}
	groupInfo.BanUID = newBanList

	err = groupRepo.UpdateGroup(ctx, groupInfo)
	if err != nil {
		log.Error("群管理解封ban人，更新数据库记录错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	go reFlashGroupInfo(ctx, req.GroupId)
	la_rsp.Success(ctx, nil)
}

type addAdminReq struct {
	GroupId   string `json:"group_id"`
	MemberUID string `json:"member_uid"`
}

// AddAdmin 群主操作，添加管理员
func AddAdmin(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &addAdminReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.GroupId)
	if err != nil {
		log.Error("群主添加管理员，从数据库获取群信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//群不存在返回错误
	if groupInfo == nil || groupInfo.Id.IsZero() || groupInfo.State != 0 {
		la_rsp.Failed(ctx, la_rsp.GroupNotExist, nil)
		return
	}

	isMember := false
	//不是本群人员不行
	for i := 0; i < len(groupInfo.UIDList); i++ {
		if groupInfo.UIDList[i] == req.MemberUID {
			isMember = true
		}
	}
	if !isMember {
		la_rsp.Failed(ctx, la_rsp.UserNotExist, nil)
		return
	}

	//已经是管理员，直接返回
	for i := 0; i < len(groupInfo.Admin); i++ {
		if groupInfo.Admin[i] == req.MemberUID {
			la_rsp.Success(ctx, nil)
			return
		}
	}

	groupInfo.Admin = append(groupInfo.Admin, req.MemberUID)
	err = groupRepo.UpdateGroup(ctx, groupInfo)
	if err != nil {
		log.Error(fmt.Sprintf("群主添加管理员，更新数据库中群信息错误：%+v", err))
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, nil)
	go reFlashGroupInfo(ctx, req.GroupId)
}

type delAdminReq struct {
	GroupId   string `json:"group_id" form:"group_id"`
	MemberUID string `json:"member_uid" form:"member_uid"`
}

// DelAdmin 群主操作，删除管理员
func DelAdmin(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	req := &delAdminReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.GroupId)
	if err != nil {
		log.Error("群主删除管理员，从数据库获取群信息错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//群不存在返回错误
	if groupInfo == nil || groupInfo.Id.IsZero() || groupInfo.State != 0 {
		la_rsp.Failed(ctx, la_rsp.GroupNotExist, nil)
		return
	}

	newAdminList := []string{}
	isExist := false
	for i := 0; i < len(groupInfo.Admin); i++ {
		if groupInfo.Admin[i] != req.MemberUID {
			newAdminList = append(newAdminList, groupInfo.Admin[i])
		} else {
			isExist = true
		}
	}
	//不是管理员，直接返回成功
	if !isExist {
		la_rsp.Success(ctx, nil)
		return
	}

	groupInfo.Admin = newAdminList
	err = groupRepo.UpdateGroup(ctx, groupInfo)
	if err != nil {
		log.Error(fmt.Sprintf("群主删除管理员，更新数据库中群信息错误：%+v", err))
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, nil)
	go reFlashGroupInfo(ctx, req.GroupId)
}

type getGroupListReq struct {
	SearchStr  string `json:"search_str" form:"search_str"`
	SearchType string `json:"search_type" form:"search_type"`
	Page       int64  `json:"page" form:"page"`
	PageSize   int64  `json:"page_size" form:"page_size"`
}

// GetGroupList 群搜索
func GetGroupList(ctx *gin.Context) {
	req := &getGroupListReq{}
	err := ctx.ShouldBind(req)
	if err != nil || req.SearchType == "" || req.SearchStr == "" || req.Page < 0 || req.PageSize < 0 {
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.Page <= 0 {
		req.Page = 0
	} else {
		req.Page -= 1
	}
	if req.SearchType == "group_id" {
		groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, req.SearchStr)
		if err != nil && err != qmgo.ErrNoSuchDocuments {
			log.Error("群ID搜索，查找数据库错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		la_rsp.Success(ctx, map[string]interface{}{
			"group_list": []*MongoModel.Group{groupInfo},
			"count":      1,
		})
		return
	}
	if req.SearchType == "group_name" {
		groupList, count, err := groupRepo.GetGroupListByGroupName(ctx, req.SearchStr, req.Page, req.PageSize)
		if err != nil && err != qmgo.ErrNoSuchDocuments {
			log.Error("群名搜索，查找数据库错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		la_rsp.Success(ctx, map[string]interface{}{
			"group_list": groupList,
			"count":      count,
		})
		return
	}
	return
}

// 刷新缓存
func reFlashGroupInfo(ctx *gin.Context, groupId string) (*MongoModel.Group, error) {
	groupInfo, err := groupRepo.GetGroupInfoByGroupId(ctx, groupId)
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		log.Error("获取群信息请求，查询数据库失败：", err)
		return nil, err
	}
	if groupInfo == nil || groupInfo.Id.IsZero() {
		return nil, nil
	}
	err = db.DB.Rdb.Del(ctx, "group_members:"+groupInfo.GroupID).Err()
	if err != nil {
		log.Error("获取群信息请求，删除redis中群成员列表失败：", err)
		return nil, err
	}
	err = db.DB.Rdb.SAdd(ctx, "group_members:"+groupInfo.GroupID, groupInfo.UIDList).Err()
	if err != nil {
		log.Error("获取群信息请求，更新redis中群成员列表失败：", err)
		return nil, err
	}
	err = db.DB.Rdb.Del(ctx, "group_bans:"+groupInfo.GroupID).Err()
	if err != nil {
		log.Error("获取群信息请求，删除redis中群被ban成员列表失败：", err)
		return nil, err
	}
	if len(groupInfo.BanUID) != 0 {
		err = db.DB.Rdb.SAdd(ctx, "group_bans:"+groupInfo.GroupID, groupInfo.BanUID).Err()
		if err != nil {
			log.Error("获取群信息请求，更新redis中群被ban成员列表失败：", err)
			return nil, err
		}
	}
	err = db.DB.Rdb.Del(ctx, "group_admin:"+groupInfo.GroupID).Err()
	if err != nil {
		log.Error("获取群信息请求，删除redis中群管理成员列表失败：", err)
		return nil, err
	}
	err = db.DB.Rdb.SAdd(ctx, "group_admin:"+groupInfo.GroupID, groupInfo.Admin).Err()
	if err != nil {
		log.Error("获取群信息请求，更新redis中群管理成员列表失败：", err)
		return nil, err
	}
	return groupInfo, nil
}
