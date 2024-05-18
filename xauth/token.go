package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	rawerrors "errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	vov2 "github.com/yituoshiniao/kit/xauth/vo"
	"github.com/yituoshiniao/kit/xcookie"
	"github.com/yituoshiniao/kit/xlog"
)

const (
	HeaderTokenName1 = "auth-token"
	HeaderTokenName2 = "token"
	CookieTokenName1 = "AUTH_TOKEN"
	CookieTokenName2 = "auth_token"
	CookieTokenName3 = "cookie"
	Guest            = -1

	HeaderAuthorize = "authorization"
)

var (
	guest = Auth{
		UserRole: Guest,
	}

	ErrGuest        = rawerrors.New("未登录")
	ErrRuleNotMatch = rawerrors.New("身份不匹配")
	ErrPermission   = rawerrors.New("未登录或身份不匹配")
)

func strPad(raw string, lens int, pasStr string) string {

	switch {
	case len(raw) >= lens:
		return raw
	case len(raw) < lens:
		return fmt.Sprintf("%s%s", raw, strings.Repeat(pasStr, lens-len(raw)))
	default:
		return raw
	}
}

type Auth struct {
	DeviceId int         `json:"device_id"`
	UserId   interface{} `json:"user_id"`
	UserRole int         `json:"usertype"`
	AppType  int         `json:"app_type"`
	Ct       int         `json:"ct"`
	Salt     string      `json:"salt"`
}

func (a Auth) GetUserId() (int, error) {
	switch v := a.UserId.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case nil:
		return 0, nil
	case string:
		c, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return c, err
	default:
		return 0, errors.New(fmt.Sprintf("conversion to int from %T not supported", v))
	}
}

func (a *Auth) StudentNumber() (int64, error) {
	uid, err := a.GetUserId()
	if err != nil {
		return 0, err
	}
	if uid == 0 {
		return 0, ErrGuest
	}
	if a.UserRole != int(vov2.USER_ROLE_STUDENT) {
		return 0, ErrRuleNotMatch
	}

	return EncodeUserNumber(uid)
}

func (a *Auth) MasterTeacherNumber() (int64, error) {
	uid, err := a.GetUserId()
	if err != nil {
		return 0, err
	}
	if uid == 0 {
		return 0, errors.WithStack(ErrGuest)
	}
	if a.UserRole != int(vov2.USER_ROLE_MASTER_TEACHER) {
		return 0, errors.WithStack(ErrRuleNotMatch)
	}

	return EncodeUserNumber(uid)
}

func (a *Auth) String() string {
	return fmt.Sprintf("%d::%d::%d::%d::%d::%s", a.UserId, a.UserRole, a.AppType, a.DeviceId, a.Ct, a.Salt)
}

func FromTokenForTest(token string) (Auth, error) {
	if len(token) < 9 {
		return guest, ErrGuest
	}

	token = token[8:]
	auth := Auth{}
	err := json.Unmarshal([]byte(strDecode(token)), &auth)
	if err != nil {
		return auth, errors.WithStack(err)
	}

	return auth, nil
}

func FromToken(ctx context.Context, token string) (Auth, error) {
	if token == "" {
		return guest, ErrGuest
	}
	auth := Auth{}
	err := json.Unmarshal([]byte(strDecode(token)), &auth)
	if err == nil {
		return auth, nil
	} else {
		xlog.Ctx(ctx).Errorw("解析prod token失败", "token", token, "err", errors.WithStack(err))
	}

	return FromTokenForTest(token)
}

func CreateTokenTest(ctx context.Context, userId int, userType int, appType int) string {
	raw := make(map[string]interface{})
	salt := RandString(8)
	raw["device_id"] = 0
	raw["user_id"] = userId
	raw["usertype"] = userType
	raw["app_type"] = appType
	raw["ct"] = time.Now().Unix()
	raw["salt"] = salt
	rawJson, err := json.Marshal(raw)
	if err != nil {
		xlog.Ctx(ctx).Infow("序列化失败")
		return ""
	}
	return salt + strEncode(string(rawJson))
}

func CreateTokenProd(ctx context.Context, userId int, userType int, appType int) string {
	raw := make(map[string]interface{})
	salt := RandString(8)
	raw["device_id"] = 0
	raw["user_id"] = userId
	raw["usertype"] = userType
	raw["app_type"] = appType
	raw["ct"] = time.Now().Unix()
	raw["salt"] = salt
	rawJson, err := json.Marshal(raw)
	if err != nil {
		xlog.Ctx(ctx).Infow("序列化失败")
		return ""
	}
	return strEncode(string(rawJson))
}

func strEncode(str string) string {
	lens := len(str)
	if lens == 0 {
		return ""
	}
	rand.Seed(time.Now().Unix())
	factor := rand.Intn(int(math.Min(math.Ceil(float64(lens/3)), 255)-1)) + 1
	c := factor % 8
	slice := strSplit(str, int(factor))
	lenSlice := len(slice)

	newSlice := make([]string, lenSlice)
	for i := 0; i < lenSlice; i++ {
		bytes := make([]byte, 0)
		tmpLenSlice := len(slice[i])
		for j := 0; j < tmpLenSlice; j++ {
			chr := slice[i][j] + uint8(c) + uint8(i)
			bytes = append(bytes, chr)
		}
		newSlice[i] = string(bytes)
	}
	ret := string(rune(factor)) + strings.Join(newSlice, "")
	return strings.TrimRight(strings.ReplaceAll(strings.ReplaceAll(base64.RawStdEncoding.EncodeToString([]byte(ret)), "+", "-"), "/", "_"), "=")
}

func strDecode(str string) string {
	lens := len(str)
	if lens == 0 {
		return ""
	}
	pad := strings.Replace(str, "-", "+", -1)
	padNew := strings.Replace(pad, "_", "/", -1)
	padLength := lens % 4
	strP := strPad(padNew, padLength, "=")
	dst, _ := base64.RawStdEncoding.DecodeString(strP)
	if len(dst) == 0 {
		return ""
	}

	factor := dst[0]
	c := factor % 8
	entity := dst[1:]
	slice := strSplit(string(entity), int(factor))
	lenSlice := len(slice)
	if lenSlice == 0 {
		return ""
	}

	newSlice := make([]string, lenSlice)
	for i := 0; i < lenSlice; i++ {
		bytes := make([]byte, 0)
		tmpLenSlice := len(slice[i])
		for j := 0; j < tmpLenSlice; j++ {
			chr := slice[i][j] - uint8(c) - uint8(i)
			bytes = append(bytes, chr)

		}
		newSlice[i] = string(bytes)
	}

	return strings.Join(newSlice, "")
}

func strSplit(str string, length int) []string {

	res := make([]string, 0)
	if length <= 0 {
		return res
	}

	start := 0
	lens := len(str)

	for {
		if start >= lens {
			break
		}

		if start+length < lens {
			res = append(res, str[start:start+length])
		} else {
			res = append(res, str[start:])
		}

		start += length

	}

	return res
}

func StudentNumberFromCtx(ctx context.Context) (int64, error) {
	auth, err := FromCtx(ctx)
	if err != nil {
		return 0, err
	}
	return auth.StudentNumber()
}

// FromCtx returns the incoming metadata in ctx if it exists.  The
// returned MD should not be modified. Writing to it may cause races.
// Modification should be made to copies of the returned MD.
func FromCtx(ctx context.Context) (auth Auth, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return guest, nil
	}

	// 优先取cookie
	cookies := xcookie.ReadCookies(md.Get(CookieTokenName3), CookieTokenName1)
	if len(cookies) > 0 && cookies[0].Value != "null" {
		return FromToken(ctx, cookies[0].Value)
	}

	// 次之读取header
	v := md.Get(HeaderTokenName1)
	if len(v) == 0 {
		return guest, nil
	}
	return FromToken(ctx, v[0])
}

func CheckPermission(err error) error {
	err = errors.Cause(err)

	if err == ErrGuest || err == ErrRuleNotMatch {
		return errors.WithStack(ErrPermission)
	}
	return nil
}

func RandString(length int) string {
	str := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghicklmnopqrstuvwxyz1234567890"
	strLen := len(str)
	if length <= 0 {
		length = 6
	}
	ret := ""
	rand.Seed(time.Now().Unix())
	for i := 0; i < length; i++ {
		ret += string(str[rand.Intn(strLen)])
	}
	return ret
}
