package user

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qiniu/qmgo"
	"lalala_im/data/Repo/Friend"
	"lalala_im/data/Repo/User"
	"lalala_im/data/entity"
	"lalala_im/data/model/MongoModel"
	"lalala_im/pkg/la_auth"
	log "lalala_im/pkg/la_log"
	"lalala_im/pkg/la_rsp"
	"lalala_im/servers/receive_server/internal/db"
	"lalala_im/servers/receive_server/internal/http/middle"
	"net/http"
	"time"
)

var (
	userRepo   = User.NewIUserRepo(db.DB.MongoDB)
	friendRepo = Friend.NewIFriendRepo(db.DB.MongoDB)
)

// LoginTest 测试登陆token是否有效
func LoginTest(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	la_rsp.Success(ctx, map[string]interface{}{
		"user": user,
	})
}

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"Email"`
	Password string `json:"password"`
	Code     string `json:"verify_code"`
}

// Register 用户注册，往用户邮箱发验证码
func Register(ctx *gin.Context) {
	req := &registerReq{}
	err := ctx.ShouldBind(req)
	if err != nil || len(req.Email) < 5 || len(req.Password) < 6 || len(req.Code) != 6 {
		log.Error("注册请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	//验证验证码
	code, err := db.DB.Rdb.Get(ctx, "verify_code:"+req.Email).Result()
	if err != nil {
		log.Error("注册请求，从redis获取验证码失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if code != req.Code {
		la_rsp.Failed(ctx, la_rsp.VerifyCodeError, nil)
		return
	}

	//	redis获取邮箱是否被注册
	temps, err := db.DB.Rdb.Get(ctx, "register:"+req.Email).Result()
	if err != nil && err != redis.Nil {
		log.Error("注册请求，请求redis错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if temps != "" {
		la_rsp.Failed(ctx, la_rsp.UserEmailExist, nil)
		return
	}

	//查库看email是否已被使用
	user, err := userRepo.GetUserInfoByEmail(ctx, req.Email)
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		log.Error("注册请求，查找数据库错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	if user != nil && !user.Id.IsZero() {
		err = db.DB.Rdb.Set(ctx, "register:"+user.Email, "exist", time.Minute*5).Err()
		log.Error("注册请求，已存在的用户写入redis缓存失败：", err)
		la_rsp.Failed(ctx, la_rsp.UserEmailExist, nil)
		return
	}

	uid := uuid.NewString()

	//使用uuid作为uid,确保不出现uid重复
	for {
		user, err := userRepo.GetUserInfoByUID(ctx, uid)
		if err != nil && err != qmgo.ErrNoSuchDocuments {
			log.Error("注册请求，数据库查询错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		if user == nil || user.Id.IsZero() {
			break
		}
	}

	//求密码的md5值，不能明文保存密码，懒得加密只保存hash值
	pw := la_auth.EncodePassword(req.Password)

	user = &MongoModel.User{
		UID:         uid,
		NickName:    req.Name,
		FaceURL:     "",
		Gender:      0,
		Email:       req.Email,
		PhoneNumber: "",
		Birth:       time.Now(),
		Status:      0,
		Password:    pw,
		Version:     0,
	}

	err = userRepo.CreateUser(ctx, user)
	if err != nil {
		log.Error("注册请求，数据库创建记录错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	token, err := la_auth.NewToken(uid, 0, 30)
	if err != nil {
		log.Error("注册请求，获取token失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	la_rsp.Success(ctx, map[string]interface{}{
		"token": token,
		"user":  entity.TransformFromModel(user),
	})
}

type loginReq struct {
	UID      string `json:"uid"`
	Email    string `json:"Email"`
	Password string `json:"password"`
}

func Login(ctx *gin.Context) {
	req := &loginReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		log.Error("登陆请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	if len(req.UID) == 0 && len(req.Email) < 5 {
		log.Error("登陆请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, nil)
		return
	}
	var userInfo *MongoModel.User
	if len(req.UID) != 0 {
		userInfo, err = userRepo.GetUserInfoByUID(ctx, req.UID)
		if err != nil && err != qmgo.ErrNoSuchDocuments {
			log.Error("登陆请求，数据库查询错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	}
	//如果没找到用邮箱找
	if userInfo == nil || userInfo.Id.IsZero() {
		if len(req.Email) > 5 {
			userInfo, err = userRepo.GetUserInfoByEmail(ctx, req.Email)
			if err != nil && err != qmgo.ErrNoSuchDocuments {
				log.Error("登陆请求，数据库查询错误：", err)
				la_rsp.Failed(ctx, la_rsp.ERROR, err)
				return
			}
		} else {
			la_rsp.Failed(ctx, la_rsp.UserNotExist, nil)
			return
		}
	}

	pw := la_auth.EncodePassword(req.Password)
	if userInfo.Password != pw {
		la_rsp.Failed(ctx, la_rsp.UserPasswordError, nil)
		return
	}

	//缓存用户信息鉴权用
	userJsonBytes, err := json.Marshal(entity.TransformFromModel(userInfo))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		log.Error("用户鉴权，用户数据解析到json错误", err)
		ctx.Abort()
		return
	}
	err = db.DB.Rdb.Set(ctx, "auth:"+userInfo.UID, string(userJsonBytes), 3600*time.Second).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		log.Error("用户鉴权，用户信息存入redis缓存错误", err)
		ctx.Abort()
		return
	}

	token, err := la_auth.NewToken(userInfo.UID, userInfo.Version, 30)
	if err != nil {
		log.Error("登陆请求，生成token失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, map[string]interface{}{
		"token": token,
		"user":  entity.TransformFromModel(userInfo),
	})
}

func GetUserInfo(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	la_rsp.Success(ctx, map[string]interface{}{
		"user": user,
	})
	return
}

type getUserShowInfoReq struct {
	UID string `json:"uid" form:"uid"`
}

func GetUserShowInfo(ctx *gin.Context) {
	req := &getUserShowInfoReq{}
	err := ctx.ShouldBind(req)
	if err != nil {
		log.Error("获取用户展示信息请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	if len(req.UID) == 0 {
		log.Error("获取用户展示信息请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, nil)
		return
	}

	userJson, err := db.DB.Rdb.Get(ctx, "auth:"+req.UID).Result()
	if err != nil && err != redis.Nil {
		log.Error("获取用户展示信息请求，redis取数据错误", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	user := &entity.User{}
	if userJson != "" {
		err = json.Unmarshal([]byte(userJson), user)
		if err != nil {
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			log.Error(fmt.Sprintf("获取用户展示信息请求，redis数据json解析错误:%+v", err))
			return
		}
		la_rsp.Success(ctx, map[string]interface{}{
			"uid":       user.UID,
			"nick_name": user.NickName,
			"face_url":  user.FaceURL,
			"gender":    user.Gender,
			"birth":     user.Birth,
			"status":    user.Status,
		})
		return
	}
	modelUser, err := userRepo.GetUserInfoByUID(ctx, req.UID)
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		log.Error(fmt.Sprintf("获取用户展示信息请求,查数据库失败：%+v", err))
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	} else if err == qmgo.ErrNoSuchDocuments {
		log.Error(fmt.Sprintf("获取用户展示信息请求,查数据库失败：%+v", err))
		la_rsp.Failed(ctx, la_rsp.UserNotExist, err)
		return
	}

	la_rsp.Success(ctx, map[string]interface{}{
		"uid":       modelUser.UID,
		"nick_name": modelUser.NickName,
		"face_url":  modelUser.FaceURL,
		"gender":    modelUser.Gender,
		"birth":     modelUser.Birth.Unix(),
		"status":    modelUser.Status,
	})
	return
}

type changePassWordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	VerifyCode  string `json:"verify_code"`
}

func ChangePassWord(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}
	req := &changePassWordReq{}
	err := ctx.ShouldBind(req)
	if err != nil || len(req.OldPassword) == 0 || len(req.NewPassword) == 0 || len(req.VerifyCode) == 0 {
		log.Error("修改密码请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	code, err := db.DB.Rdb.Get(ctx, "verify_code:"+user.Email).Result()
	if err != nil && err != redis.Nil {
		log.Error("修改密码请求，从redis获取验证码失败：", err)
		la_rsp.Failed(ctx, la_rsp.VerifyCodeError, err)
		return
	}
	if code != req.VerifyCode {
		la_rsp.Failed(ctx, la_rsp.VerifyCodeError, nil)
		return
	}
	userInfo, err := userRepo.GetUserInfoByUID(ctx, user.UID)
	if err != nil {
		log.Error("修改密码请求，从数据库获取用户信息失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	if userInfo.Password != la_auth.EncodePassword(req.OldPassword) {
		la_rsp.Failed(ctx, la_rsp.PasswordError, nil)
		return
	}
	userInfo.Password = la_auth.EncodePassword(req.NewPassword)
	userInfo.Version += 1

	//删除之前版本的用户信息
	err = db.DB.Rdb.Del(ctx, "auth:"+user.UID).Err()
	if err != nil {
		log.Error("修改密码请求，删除缓存中的用户信息失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	err = userRepo.UpdateUser(ctx, userInfo.Id.Hex(), userInfo)
	if err != nil {
		log.Error("修改密码请求，更新数据库用户信息失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	la_rsp.Success(ctx, nil)
}

type findBackPassWordReq struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	VerifyCode  string `json:"verify_code"`
}

func FindBackPassWord(ctx *gin.Context) {
	req := &findBackPassWordReq{}
	err := ctx.ShouldBind(req)
	if err != nil || len(req.OldPassword) == 0 || len(req.NewPassword) == 0 || len(req.VerifyCode) == 0 || len(req.Email) == 0 {
		log.Error("找回密码请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	userInfo, err := userRepo.GetUserInfoByEmail(ctx, req.Email)
	if err != nil {
		log.Error("修改密码请求，从数据库获取用户信息失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	//先判断验证码是否正确
	code, err := db.DB.Rdb.Get(ctx, "verify_code:"+userInfo.Email).Result()
	if err != nil && err != redis.Nil {
		log.Error("找回密码请求，从redis获取验证码失败：", err)
		la_rsp.Failed(ctx, la_rsp.VerifyCodeError, err)
		return
	}
	if code != req.VerifyCode {
		la_rsp.Failed(ctx, la_rsp.VerifyCodeError, nil)
		return
	}

	//再判断旧密码是否正确
	if userInfo.Password != la_auth.EncodePassword(req.OldPassword) {
		la_rsp.Failed(ctx, la_rsp.PasswordError, nil)
		return
	}

	userInfo.Password = la_auth.EncodePassword(req.NewPassword)
	userInfo.Version += 1

	//删除之前版本的用户信息
	err = db.DB.Rdb.Del(ctx, "auth:"+userInfo.UID).Err()
	if err != nil {
		log.Error("找回密码请求，删除缓存中的用户信息失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	err = userRepo.UpdateUser(ctx, userInfo.Id.Hex(), userInfo)
	if err != nil {
		log.Error("找回密码请求，更新数据库用户信息失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}

	la_rsp.Success(ctx, nil)
}

type changeUserDataReq struct {
	NickName    string `bson:"nick_name" json:"nick_name"`
	FaceURL     string `bson:"face_url" json:"face_url"`
	Gender      int32  `bson:"gender" json:"gender"`
	PhoneNumber string `bson:"phone_number" json:"phone_number"`
	Birth       int64  `bson:"birth" json:"birth"`
}

func ChangeUserData(ctx *gin.Context) {
	req := &changeUserDataReq{}
	err := ctx.ShouldBind(req)
	if err != nil || req.Birth > 4100428799 || req.Birth < 0 {
		log.Error("修改用户数据请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	userInfo := &MongoModel.User{}
	user := middle.GetUserFromCtx(ctx)
	if user != nil {
		userInfo, err = userRepo.GetUserInfoByObjectId(ctx, user.Id)
		if err != nil {
			log.Error("修改用户数据请求，数据库查询错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
	} else {
		log.Error("修改用户数据请求，鉴权中获取用户数据失败")
		la_rsp.Failed(ctx, la_rsp.ERROR, nil)
		return
	}
	userInfo.NickName = req.NickName
	userInfo.FaceURL = req.FaceURL
	userInfo.Gender = req.Gender
	userInfo.PhoneNumber = req.PhoneNumber
	userInfo.Birth = time.Unix(req.Birth, 0)
	err = userRepo.UpdateUser(ctx, userInfo.Id.Hex(), userInfo)
	if err != nil {
		log.Error("修改用户数据请求，数据库更新错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, entity.TransformFromModel(userInfo))
}

type getFaceImgReq struct {
	Uid string `json:"uid" form:"uid"`
}

//func GetFaceUrl(ctx *gin.Context) {
//	req := &getFaceImgReq{}
//	err := ctx.ShouldBind(req)
//	if err != nil || req.Uid == "" {
//		log.Error("请求头像，参数校验错误：", err)
//		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
//		return
//	}
//	err = db.DB.AliyunOss.GetObjectToFile("user/"+req.Uid+"/face_img/0.jpg", "./"+req.Uid+"_face_img_0.jpg")
//	if err != nil {
//		err = db.DB.AliyunOss.GetObjectToFile("default/face_img/0.jpg", "./"+req.Uid+"_face_img_0.jpg")
//		if err != nil {
//			log.Error("请求头像，创建访问url错误：", err)
//			la_rsp.Failed(ctx, la_rsp.ERROR, err)
//			return
//		}
//	}
//	ctx.File("./" + req.Uid + "_face_img_0.jpg")
//	os.Remove("./" + req.Uid + "_face_img_0.jpg")
//	return
//}

func GetFaceUrl(ctx *gin.Context) {
	req := &getFaceImgReq{}
	err := ctx.ShouldBind(req)
	if err != nil || req.Uid == "" {
		log.Error("请求头像，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	isExist, err := db.DB.AliyunOss.IsObjectExist("user/" + req.Uid + "/face_img/0.jpg")
	if err != nil {
		log.Error("请求头像，判断头像是否存在错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	//不存在则返回默认头像
	if !isExist {
		fUrl, err := db.DB.AliyunOss.SignURL("default/face_img/0.jpg", oss.HTTPGet, 60*60*24*7)
		if err != nil {
			log.Error("请求头像，创建访问默认头像url错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		la_rsp.Success(ctx, fUrl)
		return
	}
	fUrl, err := db.DB.AliyunOss.SignURL("user/"+req.Uid+"/face_img/0.jpg", oss.HTTPGet, 60*60*24*7)
	if err != nil {
		log.Error("请求头像，创建访问url错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, fUrl)
	return
}

func UpdateFaceUrl(ctx *gin.Context) {
	user := middle.GetUserFromCtx(ctx)
	if user == nil {
		la_rsp.Failed(ctx, la_rsp.ERROR, errors.New("从上下文获取用户信息失败"))
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil || len(form.File["face_img"]) == 0 {
		log.Error("更换头像，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	tempFile := form.File["face_img"][0]
	faceFile, err := tempFile.Open()
	if err != nil || len(form.File["face_img"]) == 0 {
		log.Error("更换头像，获取文件错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}

	err = db.DB.AliyunOss.PutObject("user/"+user.UID+"/face_img/0.jpg", faceFile)
	if err != nil {
		log.Error("更换头像，上传文件到阿里云错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	fUrl, err := db.DB.AliyunOss.SignURL("user/"+user.UID+"/face_img/0.jpg", oss.HTTPGet, 60*60*24*7)
	if err != nil {
		log.Error("更换头像，创建访问url错误：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, fUrl)
}

type getUserListReq struct {
	QueryStr  string `json:"query_str" form:"query_str"`
	QueryType string `json:"query_type" form:"query_type"`
	PageSize  int64  `json:"page_size" form:"page_size"`
	Page      int64  `json:"page" form:"page"`
}

func GetUserList(ctx *gin.Context) {
	req := &getUserListReq{}
	err := ctx.ShouldBind(req)
	if err != nil || req.QueryStr == "" || req.QueryType == "" || req.PageSize < 0 || req.Page < 0 {
		log.Error("获取用户列表，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.QueryType == "nick_name" {
		usersList, err := userRepo.GetUserListByNickName(ctx, req.QueryStr, req.PageSize, req.Page)
		if err != nil {
			log.Error("以昵称获取用户列表，数据库查询错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		results := []*entity.User{}
		for _, user := range usersList {
			results = append(results, entity.TransformFromModel(user))
		}
		la_rsp.Success(ctx, results)
		return
	} else if req.QueryType == "Email" {
		user, err := userRepo.GetUserInfoByEmail(ctx, req.QueryStr)
		if err != nil {
			log.Error("以邮箱获取用户列表，数据库查询错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		la_rsp.Success(ctx, []*entity.User{entity.TransformFromModel(user)})
		return
	} else if req.QueryType == "uid" {
		user, err := userRepo.GetUserInfoByUID(ctx, req.QueryStr)
		if err != nil {
			log.Error("以uid获取用户列表，数据库查询错误：", err)
			la_rsp.Failed(ctx, la_rsp.ERROR, err)
			return
		}
		la_rsp.Success(ctx, []*entity.User{entity.TransformFromModel(user)})
		return
	}
}
