package services

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	
	"strings"
	"time"
)

// RedisService emulates a Redis server that responds to common
// Redis commands with plausible data, including fake sensitive keys.
type RedisService struct {
	BaseService
}

func NewRedisService() *RedisService {
	return &RedisService{}
}

func (s *RedisService) Name() string { return "redis" }

func (s *RedisService) HandleConn(ctx *SessionContext) error {
	if ctx.Conn == nil {
		return errors.New("redis service requires a TCP connection")
	}

	_ = ctx.Conn.SetDeadline(time.Now().Add(ctx.Deadline))
	reader := bufio.NewReader(ctx.Conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "redis.command", map[string]any{
			"command": command,
		})

		response := handleRedisCommand(command)
		if _, err := fmt.Fprint(ctx.Conn, response); err != nil {
			return err
		}

		if strings.EqualFold(command, "QUIT") {
			return nil
		}
	}
}

func (s *RedisService) HandlePacket(*PacketContext) error {
	return nil
}

// handleRedisCommand returns plausible Redis protocol responses
func handleRedisCommand(cmd string) string {
	upper := strings.ToUpper(strings.TrimSpace(cmd))

	// RESP protocol responses
	switch {
	case strings.HasPrefix(upper, "PING"):
		return "+PONG\r\n"
	case strings.HasPrefix(upper, "INFO"):
		return redisInfo()
	case strings.HasPrefix(upper, "KEYS *"), strings.HasPrefix(upper, "KEYS *"):
		return redisKeys()
	case strings.HasPrefix(upper, "DBSIZE"):
		return ":847\r\n" // 847 keys — looks like real usage
	case strings.HasPrefix(upper, "GET "):
		return redisGet(strings.TrimPrefix(upper, "GET "))
	case strings.HasPrefix(upper, "SET "):
		return "+OK\r\n"
	case strings.HasPrefix(upper, "AUTH "):
		return "+OK\r\n" // Accept any password
	case strings.HasPrefix(upper, "SELECT "):
		return "+OK\r\n"
	case strings.HasPrefix(upper, "CONFIG GET "):
		return redisConfig()
	case strings.HasPrefix(upper, "CLIENT LIST"):
		return redisClientList()
	case strings.HasPrefix(upper, "FLUSHALL"), strings.HasPrefix(upper, "FLUSHDB"):
		return "+OK\r\n"
	case upper == "QUIT":
		return "+OK\r\n"
	default:
		return "-ERR unknown command '" + cmd + "'\r\n"
	}
}

func redisInfo() string {
	return "$" + fmt.Sprintf("%d", len(redisInfoString)) + "\r\n" + redisInfoString + "\r\n"
}

var redisInfoString = `# Server
redis_version:7.2.3
redis_mode:standalone
os:Linux 6.1.0-17-amd64
tcp_port:6379
uptime_in_seconds:864000
uptime_in_days:10

# Clients
connected_clients:12
blocked_clients:2

# Memory
used_memory:847296
used_memory_human:828.41K
maxmemory:0

# Keyspace
db0:keys=847,expires=23,avg_ttl=3600000`

func redisKeys() string {
	// Return tempting key names
	keys := []string{
		"session:admin:token",
		"config:database:url",
		"cache:users:active",
		"secret:api:key:production",
		"backup:s3:credentials",
		"auth:ldap:bind_password",
		"deploy:ssh:private_key",
		"monitoring:grafana:admin",
		"payment:stripe:webhook_secret",
		"internal:vpn:config",
	}
	result := "*" + fmt.Sprintf("%d", len(keys)) + "\r\n"
	for _, key := range keys {
		result += "$" + fmt.Sprintf("%d", len(key)) + "\r\n" + key + "\r\n"
	}
	return result
}

func redisGet(key string) string {
	// Return plausible values for tempting keys
	values := map[string]string{
		"session:admin:token":       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.honeytrap",
		"config:database:url":      "postgres://admin:s3cret@db.internal:5432/operations",
		"secret:api:key:production": "sk-proj-htk-REDACTED-DECOY",
		"backup:s3:credentials":    "AKIAHTKDECOY/wJalrXUtnFEMI/DECOY",
		"auth:ldap:bind_password":  "BindPassword123!",
	}
	if val, ok := values[strings.TrimSpace(key)]; ok {
		return "$" + fmt.Sprintf("%d", len(val)) + "\r\n" + val + "\r\n"
	}
	return "$-1\r\n" // Nil
}

func redisConfig() string {
	return "*2\r\n$4\r\ndbfilename\r\n$10\r\ndump.rdb\r\n"
}

func redisClientList() string {
	return "$" + fmt.Sprintf("%d", len(redisClientListString)) + "\r\n" + redisClientListString + "\r\n"
}

var redisClientListString = `id=3 addr=10.0.1.15:52341 fd=7 name=monitoring age=864 idle=0 flags=N db=0 cmd=info
id=5 addr=10.0.1.22:48921 fd=9 name=app-server age=432 idle=1 flags=N db=0 cmd=get
id=7 addr=10.0.1.8:39102 fd=11 name=worker age=216 idle=3 flags=N db=0 cmd=keys`
