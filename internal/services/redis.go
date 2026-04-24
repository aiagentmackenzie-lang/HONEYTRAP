package services

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// RedisService emulates a Redis server that responds to common
// Redis commands with plausible data, including fake sensitive keys.
// It properly parses the RESP (REdis Serialization Protocol) so it
// works with redis-cli and all standard Redis client libraries.
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
		// Parse RESP protocol — real Redis clients always send RESP-encoded commands
		cmd, args, err := readRESPCommand(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			// If RESP parsing fails, try inline command (for netcat/manual testing)
			line, lineErr := reader.ReadString('\n')
			if lineErr != nil {
				if errors.Is(lineErr, io.EOF) {
					return nil
				}
				return lineErr
			}
			command := strings.TrimSpace(line)
			if command == "" {
				continue
			}
			_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "redis.command", map[string]any{
				"command": command,
				"protocol": "inline",
			})
			response := handleRedisCommand(command)
			if _, err := fmt.Fprint(ctx.Conn, response); err != nil {
				return err
			}
			if strings.EqualFold(command, "QUIT") {
				return nil
			}
			continue
		}

		// Record the full command with args
		eventPayload := map[string]any{
			"command":  cmd,
			"protocol": "resp",
		}
		if len(args) > 0 {
			eventPayload["args"] = args
		}
		_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "redis.command", eventPayload)

		response := dispatchRedisCommand(cmd, args)
		if _, err := fmt.Fprint(ctx.Conn, response); err != nil {
			return err
		}

		if strings.EqualFold(cmd, "QUIT") {
			return nil
		}
	}
}

func (s *RedisService) HandlePacket(*PacketContext) error {
	return nil
}

// ─── RESP Protocol Parser ─────────────────────────────────────────────────────

// readRESPCommand reads one RESP-encoded command from the stream.
// Returns the command name and its arguments, or an error if the stream
// doesn't look like RESP.
func readRESPCommand(reader *bufio.Reader) (string, []string, error) {
	// Peek at the first byte to see if this is RESP
	b, err := reader.ReadByte()
	if err != nil {
		return "", nil, err
	}

	if b != '*' {
		// Not RESP — put byte back and return error (caller falls back to inline)
		_ = reader.UnreadByte()
		return "", nil, fmt.Errorf("not RESP protocol")
	}

	// Read array length
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || count <= 0 {
		return "", nil, fmt.Errorf("invalid RESP array length: %q", line)
	}

	// Read each element
	elements := make([]string, 0, count)
	for i := 0; i < count; i++ {
		prefix, err := reader.ReadByte()
		if err != nil {
			return "", nil, err
		}
		if prefix != '$' {
			return "", nil, fmt.Errorf("expected '$' prefix for bulk string, got '%c'", prefix)
		}

		lenLine, err := reader.ReadString('\n')
		if err != nil {
			return "", nil, err
		}
		length, err := strconv.Atoi(strings.TrimSpace(lenLine))
		if err != nil {
			return "", nil, fmt.Errorf("invalid bulk string length: %q", lenLine)
		}
		if length < 0 {
			// Null bulk string
			elements = append(elements, "")
			continue
		}

		data := make([]byte, length+2) // +2 for \r\n
		if _, err := io.ReadFull(reader, data); err != nil {
			return "", nil, err
		}
		elements = append(elements, string(data[:length]))
	}

	if len(elements) == 0 {
		return "", nil, fmt.Errorf("empty RESP command")
	}

	return elements[0], elements[1:], nil
}

// ─── Command Dispatch ─────────────────────────────────────────────────────────

// dispatchRedisCommand handles RESP-parsed commands with separate args.
func dispatchRedisCommand(cmd string, args []string) string {
	upper := strings.ToUpper(cmd)

	switch upper {
	case "PING":
		if len(args) > 0 {
			return "$" + fmt.Sprintf("%d", len(args[0])) + "\r\n" + args[0] + "\r\n"
		}
		return "+PONG\r\n"
	case "INFO":
		return redisInfo()
	case "COMMAND":
		// redis-cli sends COMMAND DOCS on connect
		return "*0\r\n"
	case "DBSIZE":
		return ":847\r\n"
	case "KEYS":
		pattern := "*"
		if len(args) > 0 {
			pattern = args[0]
		}
		if pattern == "*" {
			return redisKeys()
		}
		return "*0\r\n"
	case "GET":
		if len(args) > 0 {
			return redisGet(args[0])
		}
		return "-ERR wrong number of arguments for 'get' command\r\n"
	case "SET":
		return "+OK\r\n"
	case "AUTH":
		return "+OK\r\n"
	case "SELECT":
		return "+OK\r\n"
	case "CONFIG":
		if len(args) > 0 && strings.ToUpper(args[0]) == "GET" {
			return redisConfig()
		}
		return "-ERR unknown subcommand for CONFIG\r\n"
	case "CLIENT":
		if len(args) > 0 && strings.ToUpper(args[0]) == "LIST" {
			return redisClientList()
		}
		return "+OK\r\n"
	case "FLUSHALL", "FLUSHDB":
		return "+OK\r\n"
	case "DEL":
		return ":1\r\n"
	case "EXISTS":
		return ":1\r\n"
	case "TYPE":
		return "+string\r\n"
	case "TTL":
		return ":-1\r\n"
	case "HELLO":
		// RESP3 handshake — respond with minimal HELLO
		return "%7\r\n$6\r\nserver\r\n$5\r\nredis\r\n$7\r\nversion\r\n$5\r\n7.2.3\r\n$5\r\nproto\r\n:2\r\n$2\r\nid\r\n:42\r\n$4\r\nmode\r\n$10\r\nstandalone\r\n"
	case "QUIT":
		return "+OK\r\n"
	default:
		return "-ERR unknown command '" + cmd + "'\r\n"
	}
}

// handleRedisCommand handles inline (newline-delimited) commands for netcat users.
func handleRedisCommand(cmd string) string {
	upper := strings.ToUpper(strings.TrimSpace(cmd))

	switch {
	case strings.HasPrefix(upper, "PING"):
		return "+PONG\r\n"
	case strings.HasPrefix(upper, "INFO"):
		return redisInfo()
	case strings.HasPrefix(upper, "KEYS *"), strings.HasPrefix(upper, "KEYS *"):
		return redisKeys()
	case strings.HasPrefix(upper, "DBSIZE"):
		return ":847\r\n"
	case strings.HasPrefix(upper, "GET "):
		return redisGet(strings.TrimPrefix(upper, "GET "))
	case strings.HasPrefix(upper, "SET "):
		return "+OK\r\n"
	case strings.HasPrefix(upper, "AUTH "):
		return "+OK\r\n"
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

// ─── Fake Data Responses ──────────────────────────────────────────────────────

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
	values := map[string]string{
		"session:admin:token":       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.0ps3c",
		"config:database:url":       "postgres://admin:s3cret@db.primary.internal.ops:5432/operations",
		"secret:api:key:production": "sk-proj-REDACTED-DECOY",
		"backup:s3:credentials":    "AKIADECOY/wJalrXUtnFEMI/DECOY",
		"auth:ldap:bind_password":   "BindPassword123!",
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