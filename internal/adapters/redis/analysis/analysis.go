// Package analysis builds a synthetic static graph of Redis commands.
//
// Redis is a running service rather than a source tree, so there is nothing to
// statically analyze the way sen does for Go or Node. Instead we synthesize
// a graph whose nodes are well-known Redis commands so the TUI has a stable
// surface onto which observed command activity can be attributed.
package analysis

import (
	"slices"

	"github.com/briheet/sen/internal/model"
)

const (
	// ModulePath is the synthetic package path returned as the graph namespace.
	ModulePath = "sen/redis"
	moduleName = "redis"
	packageID  = model.PackageID(1)
	rootID     = model.NodeID(1)
)

// knownCommands is the set of commands exposed to the TUI. Keeping it fixed
// gives the collector stable node ids across snapshots.
var knownCommands = []string{
	"APPEND", "BGSAVE", "BITCOUNT", "BITFIELD",
	"BITOP", "BLMOVE", "BLPOP", "BRPOP",
	"BRPOPLPUSH", "BZPOPMAX", "BZPOPMIN", "CLIENT",
	"CLUSTER", "COMMAND", "CONFIG", "COPY",
	"DBSIZE", "DEBUG", "DECR", "DECRBY",
	"DEL", "DISCARD", "DUMP", "ECHO",
	"EVAL", "EVALSHA", "EXEC", "EXISTS",
	"EXPIRE", "EXPIREAT", "FLUSHALL",
	"FLUSHDB", "GEOADD", "GEODIST", "GEOHASH",
	"GEOPOS", "GEORADIUS", "GET", "GETBIT",
	"GETDEL", "GETEX", "GETRANGE", "GETSET",
	"HDEL", "HEXISTS", "HGET", "HGETALL",
	"HINCRBY", "HINCRBYFLOAT", "HKEYS", "HLEN",
	"HMGET", "HMSET", "HRANDFIELD", "HSCAN",
	"HSET", "HSETNX", "HSTRLEN", "HVALS",
	"INCR", "INCRBY", "INCRBYFLOAT", "INFO",
	"KEYS", "LASTSAVE", "LATENCY", "LINDEX",
	"LINSERT", "LLEN", "LMOVE", "LMPOP",
	"LPOP", "LPOS", "LPUSH", "LPUSHX",
	"LRANGE", "LREM", "LSET", "LTRIM",
	"MGET", "MIGRATE", "MONITOR", "MOVE",
	"MSET", "MSETNX", "MULTI", "OBJECT",
	"PERSIST", "PEXPIRE", "PEXPIREAT", "PFADD",
	"PFCOUNT", "PFMERGE", "PING", "PSETEX",
	"PSUBSCRIBE", "PTTL", "PUBLISH",
	"PUBSUB", "QUIT", "RANDOMKEY", "RENAME",
	"RENAMENX", "RESTORE", "RPOP", "RPOPLPUSH",
	"RPUSH", "RPUSHX", "SADD", "SAVE",
	"SCAN", "SCARD", "SCRIPT", "SDIFF",
	"SDIFFSTORE", "SELECT", "SET", "SETBIT",
	"SETEX", "SETNX", "SETRANGE", "SHUTDOWN",
	"SINTER", "SINTERCARD", "SINTERSTORE", "SISMEMBER",
	"SLAVEOF", "SLOWLOG", "SMEMBERS", "SMISMEMBER",
	"SMOVE", "SORT", "SPOP", "SRANDMEMBER",
	"SREM", "SSCAN", "STRLEN", "SUBSCRIBE",
	"SUNION", "SUNIONSTORE", "SWAPDB",
	"SYNC", "TIME", "TOUCH", "TTL",
	"TYPE", "UNLINK", "UNSUBSCRIBE", "UNWATCH",
	"WATCH", "XACK", "XADD", "XAUTOCLAIM",
	"XCLAIM", "XDEL", "XGROUP", "XINFO",
	"XLEN", "XPENDING", "XRANGE", "XREAD",
	"XREADGROUP", "XREVRANGE", "XSETID", "XTRIM",
	"ZADD", "ZCARD", "ZCOUNT", "ZDIFF",
	"ZDIFFSTORE", "ZINCRBY", "ZINTER", "ZINTERSTORE",
	"ZLEXCOUNT", "ZMPOP", "ZPOPMAX", "ZPOPMIN",
	"ZRANDMEMBER", "ZRANGE", "ZRANGEBYLEX", "ZRANGEBYSCORE",
	"ZRANK", "ZREM", "ZREMRANGEBYLEX", "ZREMRANGEBYRANK",
	"ZREMRANGEBYSCORE", "ZREVRANGE", "ZREVRANGEBYSCORE",
	"ZREVRANK", "ZSCAN", "ZSCORE", "ZUNION",
	"ZUNIONSTORE",
}

// BuildGraph makes a synthetic graph where every command is a node inside a
// single synthetic package, with one synthetic "file" per command so the
// runtime mapper can attribute observed heat back to a specific node.
func BuildGraph() *model.StaticGraph {
	graph := &model.StaticGraph{
		Root:     rootID,
		Nodes:    make(map[model.NodeID]*model.StaticNode),
		Files:    make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	}
	graph.Packages[packageID] = &model.Package{Path: ModulePath, Name: moduleName}
	addNode(graph, rootID, "redis-server")

	for index, name := range knownCommands {
		id := model.NodeID(index + 2)
		addNode(graph, id, name)
		graph.Nodes[id].In = []model.NodeID{rootID}
		graph.Nodes[rootID].Out = append(graph.Nodes[rootID].Out, id)
	}
	return graph
}

// addNode gives each synthetic function its own synthetic file. The runtime
// mapper can then resolve a commandstats frame exactly as it resolves a source
// frame from Go or Node.js.
func addNode(graph *model.StaticGraph, id model.NodeID, name string) {
	fileID := model.FileID(id)
	graph.Nodes[id] = &model.StaticNode{
		Name: name,
		ID:   id,
		Pkg:  packageID,
		Syntax: model.Syntax{
			Kind:  model.SyntaxFuncDecl,
			File:  fileID,
			Start: model.Position{Line: 1},
			End:   model.Position{Line: 1},
		},
	}
	graph.Files[fileID] = &model.StaticFile{
		ID:        fileID,
		Path:      ModulePath + "/" + name,
		Package:   packageID,
		Functions: []model.NodeID{id},
	}
}

// IsKnownCommand reports whether name is a well-known Redis command.
func IsKnownCommand(name string) bool {
	_, found := slices.BinarySearch(knownCommands, name)
	return found
}
