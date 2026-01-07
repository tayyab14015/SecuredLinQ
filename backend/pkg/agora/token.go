package agora

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/rand"
	"sort"
	"time"
)

const (
	RolePublisher  = 1
	RoleSubscriber = 2
)

// Privilege types
const (
	PrivilegeJoinChannel        = uint16(1)
	PrivilegePublishAudioStream = uint16(2)
	PrivilegePublishVideoStream = uint16(3)
	PrivilegePublishDataStream  = uint16(4)
)

// AccessToken represents an Agora access token
type AccessToken struct {
	AppID          string
	AppCertificate string
	ChannelName    string
	UID            string
	Ts             uint32
	Salt           uint32
	Message        map[uint16]uint32
	Signature      string
	CrcChannelName uint32
	CrcUID         uint32
}

// NewAccessToken creates a new access token
func NewAccessToken(appID, appCertificate, channelName, uid string) *AccessToken {
	ts := uint32(time.Now().Unix()) + 24*3600
	rand.Seed(time.Now().UnixNano())
	salt := rand.Uint32()

	return &AccessToken{
		AppID:          appID,
		AppCertificate: appCertificate,
		ChannelName:    channelName,
		UID:            uid,
		Ts:             ts,
		Salt:           salt,
		Message:        make(map[uint16]uint32),
		CrcChannelName: crc32.ChecksumIEEE([]byte(channelName)),
		CrcUID:         crc32.ChecksumIEEE([]byte(uid)),
	}
}

// AddPrivilege adds a privilege to the token
func (token *AccessToken) AddPrivilege(privilege uint16, expireTimestamp uint32) {
	token.Message[privilege] = expireTimestamp
}

// Build builds the token string
func (token *AccessToken) Build() (string, error) {
	msg := token.packMessage()
	val := token.AppID + token.ChannelName + token.UID + string(msg)

	sig := hmacSign(token.AppCertificate, val)
	crcChannel := crc32.ChecksumIEEE([]byte(token.ChannelName))
	crcUid := crc32.ChecksumIEEE([]byte(token.UID))

	content := packContent(sig, crcChannel, crcUid, msg)

	result := "006" + token.AppID + base64.StdEncoding.EncodeToString(content)
	return result, nil
}

func (token *AccessToken) packMessage() []byte {
	var buf bytes.Buffer

	// Pack salt
	packUint32ToBuf(&buf, token.Salt)

	// Pack timestamp
	packUint32ToBuf(&buf, token.Ts)

	// Pack privileges map
	packMapUint32ToBuf(&buf, token.Message)

	return buf.Bytes()
}

func packContent(sig []byte, crcChannel, crcUid uint32, msg []byte) []byte {
	var buf bytes.Buffer

	// Pack signature
	packBytesToBuf(&buf, sig)

	// Pack crc channel name
	packUint32ToBuf(&buf, crcChannel)

	// Pack crc uid
	packUint32ToBuf(&buf, crcUid)

	// Pack message
	packBytesToBuf(&buf, msg)

	return buf.Bytes()
}

func hmacSign(key, msg string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	return h.Sum(nil)
}

func packUint16ToBuf(buf *bytes.Buffer, val uint16) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, val)
	buf.Write(b)
}

func packUint32ToBuf(buf *bytes.Buffer, val uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, val)
	buf.Write(b)
}

func packBytesToBuf(buf *bytes.Buffer, val []byte) {
	packUint16ToBuf(buf, uint16(len(val)))
	buf.Write(val)
}

func packMapUint32ToBuf(buf *bytes.Buffer, val map[uint16]uint32) {
	// Sort keys for consistent output
	keys := make([]uint16, 0, len(val))
	for k := range val {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	packUint16ToBuf(buf, uint16(len(val)))
	for _, k := range keys {
		packUint16ToBuf(buf, k)
		packUint32ToBuf(buf, val[k])
	}
}

// GenerateRTCToken generates an RTC token for video calling
func GenerateRTCToken(appID, appCertificate, channelName, uid string, role int, expireSeconds uint32) (string, error) {
	if appID == "" {
		return "", fmt.Errorf("appID is required")
	}
	if appCertificate == "" {
		return "", fmt.Errorf("appCertificate is required")
	}

	token := NewAccessToken(appID, appCertificate, channelName, uid)

	expireTimestamp := uint32(time.Now().Unix()) + expireSeconds

	// Add privileges based on role
	token.AddPrivilege(PrivilegeJoinChannel, expireTimestamp)

	if role == RolePublisher {
		token.AddPrivilege(PrivilegePublishAudioStream, expireTimestamp)
		token.AddPrivilege(PrivilegePublishVideoStream, expireTimestamp)
		token.AddPrivilege(PrivilegePublishDataStream, expireTimestamp)
	}

	return token.Build()
}
