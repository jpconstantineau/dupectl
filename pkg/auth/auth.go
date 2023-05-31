package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func getSecret() []byte {
	data := viper.GetString("server.serverid") + viper.GetString("server.apikey")
	SECRET := []byte(data)
	return SECRET
}

func getApiKey() string {
	api_key := viper.GetString("server.apikey")
	return api_key
}

func macUint64() uint64 {
	interfaces, err := net.Interfaces()
	if err != nil {
		return uint64(0)
	}

	for _, i := range interfaces {
		if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {

			// Skip locally administered addresses
			if i.HardwareAddr[0]&2 == 2 {
				continue
			}

			var mac uint64
			for j, b := range i.HardwareAddr {
				if j >= 8 {
					break
				}
				mac <<= 8
				mac += uint64(b)
			}

			return mac
		}
	}

	return uint64(0)
}

func GenerateMachineID() string {

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	mac := macUint64()

	ServerID := []byte(fmt.Sprintf("%s-%d", hostname, mac))
	encoded := base64.StdEncoding.EncodeToString(ServerID)

	return encoded
}

func DecodeMachineID(clientID string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(clientID)
	if err != nil {
		return "", err
	} else {
		decstr := string(decoded)
		return decstr, nil
	}
}

func GenerateAPISeed() string {
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	return uuid
}

func CreateJWT(h http.Header) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	if h["Clientid"] != nil {
		claims["Clientid"] = h["Clientid"][0]
	}

	if h["Uniqueid"] != nil {
		claims["Uniqueid"] = h["Uniqueid"][0]
	}

	tokenstr, err := token.SignedString(getSecret())
	if err != nil {
		fmt.Println(err.Error())
		return "", err

	}
	return tokenstr, nil
}

func GetJWT(w http.ResponseWriter, r *http.Request) {
	if r.Header["Key"] != nil {
		if r.Header["Key"][0] == getApiKey() {
			token, err := CreateJWT(r.Header)
			if err != nil {

				return
			} else {
				fmt.Fprint(w, token)
			}

		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not Authorized - Invalid key"))
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not Authorized - No key"))
	}
}

func RegisterJWT(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Key"] != nil {
			if r.Header["Key"][0] == getApiKey() {
				token, err := CreateJWT(r.Header)
				if err != nil {
					return
				} else {
					fmt.Fprint(w, token)
					next(w, r)
				}

			} else {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Not Authorized - Invalid key"))
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not Authorized - No key"))
		}

	})
}

func ValidateJWT(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(t *jwt.Token) (interface{}, error) {
				_, ok := t.Method.(*jwt.SigningMethodHMAC)
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(" Not Authorized "))
				}
				return getSecret(), nil
			})

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(" Not Authorized " + err.Error()))
			}
			if token.Valid {
				if token.Claims.(jwt.MapClaims)["Clientid"] == nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(" Clientid Missing "))
					return
				}
				if token.Claims.(jwt.MapClaims)["Uniqueid"] == nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(" Uniqueid Missing "))
					return
				}
				next(w, r)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(" Not Authorized "))
		}

	})
}
