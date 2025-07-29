package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"event-service/pkg/constants"
	"event-service/pkg/consul"
	"github.com/hashicorp/consul/api"
)

type UserService interface {
	GetUserInfor(ctx context.Context, userID string) (*UserInfor, error)
	GetAllUser(ctx context.Context) ([]*UserInfor, error)
	GetTokenUser(ctx context.Context, userID string) (*[]string, error)
}

type userService struct {
	client *callAPI
}

type callAPI struct {
	client       consul.ServiceDiscovery
	clientServer *api.CatalogService
}

var (
	mainService = "go-main-service"
)

func NewUserService(client *api.Client) UserService {
	mainServiceAPI := NewServiceAPI(client, mainService)
	return &userService{
		client: mainServiceAPI,
	}
}

func NewServiceAPI(client *api.Client, serviceName string) *callAPI {
	sd, err := consul.NewServiceDiscovery(client, serviceName)
	if err != nil {
		fmt.Printf("Error creating service discovery: %v\n", err)
		return nil
	}

	service, err := sd.DiscoverService()
	if err != nil {
		fmt.Printf("Error discovering service: %v\n", err)
		return nil
	}

	if os.Getenv("LOCAL_TEST") == "true" {
        fmt.Println("Running in LOCAL_TEST mode — overriding service address to localhost")
        service.ServiceAddress = "localhost"
    }

	return &callAPI{
		client:       sd,
		clientServer: service,
	}
}

func (u *userService) GetTokenUser(ctx context.Context, userID string) (*[]string, error) {
	return u.client.getTokenUser(ctx, userID)
}

func (u *userService) GetUserInfor(ctx context.Context, userID string) (*UserInfor, error) {

    token, ok := ctx.Value(constants.TokenKey).(string)
	if !ok {
		return nil, fmt.Errorf("token not found in context")
	}

	data, err := u.client.GetUserInfor(userID, token)
	if err != nil {
		return nil, err
	}
	
    if data == nil {
        return nil, fmt.Errorf("no user data found for userID: %s", userID)
    }

    innerData, ok := data["data"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid response format: missing 'data' field")
    }

    var roleName string
    rolesRaw, ok := innerData["roles"].([]interface{})
    if ok && len(rolesRaw) > 0 {
        firstRole, ok := rolesRaw[0].(map[string]interface{})
        if ok {
            roleName = safeString(firstRole["role_name"])
        }
    }

    return &UserInfor{
        UserID:   safeString(innerData["id"]),
        UserName: safeString(innerData["username"]),
        FullName: safeString(innerData["fullname"]),
        Avartar:  safeString(innerData["avatar"]),
        Role:     roleName,
    }, nil
}

func (u *userService) GetAllUser(ctx context.Context) ([]*UserInfor, error) {

	token, ok := ctx.Value(constants.TokenKey).(string)
	if !ok {
		return nil, fmt.Errorf("token not found in context")
	}
	
	data := u.client.GetAllUser(token)
	if data == nil {
		return nil, fmt.Errorf("no user data found")
	}

	var users []*UserInfor

	for _, user := range data {
		userInfor := &UserInfor{
			UserID:   safeString(user["id"]),
			UserName: safeString(user["username"]),
			FullName: safeString(user["fullname"]),
			Avartar:  safeString(user["avatar"]),
		}
		users = append(users, userInfor)
	}

	return users, nil
}


func safeString(val interface{}) string {
    if val == nil {
        return ""
    }
    str, ok := val.(string)
    if !ok {
        return ""
    }
    return str
}


func (c *callAPI) GetUserInfor(userID string, token string) (map[string]interface{}, error) {

	endpoint := fmt.Sprintf("/v1/user/%s", userID)

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + token,
	}

	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodGet, nil, header)
	if err != nil {
		fmt.Printf("Error calling API: %v\n", err)
		return nil, err
	}

	var userData interface{}
	
	err = json.Unmarshal([]byte(res), &userData)
	if err != nil {
		fmt.Printf("Error unmarshalling response: %v\n", err)
		return nil, err
	}


	myMap := userData.(map[string]interface{})

	return myMap, nil
}

func (c *callAPI) GetAllUser(token string) []map[string]interface{} {

	endpoint := "/v1/user/all/?role=all"

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + token,
	}

	
	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodGet, nil, header)
	if err != nil {
		fmt.Printf("Error calling API: %v\n", err)
		return nil
	}

	var parse map[string]interface{}
	if err := json.Unmarshal([]byte(res), &parse); err != nil {
		fmt.Printf("Error calling API: %v\n", err)
		return nil
	}

	dataListRaw, ok := parse["data"].([]interface{})
	if !ok {
		fmt.Printf("Error calling API: %v\n", err)
		return nil
	}

	users := make([]map[string]interface{}, 0)

	for _, item := range dataListRaw {
		users = append(users, item.(map[string]interface{}))
	}

	return users

}

func (c *callAPI) getTokenUser(ctx context.Context, userID string) (*[]string, error) {
	
	token, ok := ctx.Value(constants.TokenKey).(string)
	if !ok {
		return nil, fmt.Errorf("token not found in context")
	}

	endpoint := fmt.Sprintf("/v1/user-token-fcm/all/%s", userID)
	fmt.Printf("endpoint: %s\n", endpoint)
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + token,
	}
	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodGet, nil, header)
	if err != nil {
		fmt.Printf("Error calling API: %v\n", err)
		return nil, err
	}

	var userData map[string]interface{}
	err = json.Unmarshal([]byte(res), &userData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling: %v", err)
	}

	rawTokens, ok := userData["data"].([]interface{})
	if !ok {
		return nil, nil
	}

	tokens := make([]string, len(rawTokens))
	for i, t := range rawTokens {
		tokens[i] = fmt.Sprintf("%v", t)
	}

	return &tokens, nil

}