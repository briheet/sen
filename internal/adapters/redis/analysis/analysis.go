// Package analysis builds a synthetic static graph of Redis commands.
//
// Redis is a running service rather than a source tree, so there is nothing to
// statically analyze the way sen does for Go or Node. Instead we synthesize
// a graph whose nodes are the well-known Redis commands so the TUI has a stable
// surface onto which observed per-command heat and latency can be attributed.
package analysis

import (
	"sort"
	"sync"

	"github.com/briheet/sen/internal/model"
)

const (
	// ModulePath is the synthetic package path returned as the graph namespace.
	ModulePath = "sen/redis"
	moduleName = "redis"
)

// command describes one well-known Redis command.
type command struct {
	name string
}

// commands is the set of commands exposed to the TUI. Keeping this fixed gives
// the collector stable node ids across snapshots; observed traffic is then
// attributed to whichever of these commands ran.
var commands = []command{
	{name: "APPEND"}, {name: "BGSAVE"}, {name: "BITCOUNT"}, {name: "BITFIELD"},
	{name: "BITOP"}, {name: "BLMOVE"}, {name: "BLPOP"}, {name: "BRPOP"},
	{name: "BRPOPLPUSH"}, {name: "BZPOPMAX"}, {name: "BZPOPMIN"}, {name: "CLIENT"},
	{name: "CLUSTER"}, {name: "COMMAND"}, {name: "CONFIG"}, {name: "COPY"},
	{name: "DBSIZE"}, {name: "DEBUG"}, {name: "DECR"}, {name: "DECRBY"},
	{name: "DEL"}, {name: "DISCARD"}, {name: "DUMP"}, {name: "ECHO"},
	{name: "EVAL"}, {name: "EVALSHA"}, {name: "EXEC"}, {name: "EXISTS"},
	{name: "EXPIRE"}, {name: "EXPIREAT"}, {name: "FLUSHALL"},
	{name: "FLUSHDB"}, {name: "GEOADD"}, {name: "GEODIST"}, {name: "GEOHASH"},
	{name: "GEOPOS"}, {name: "GEORADIUS"}, {name: "GET"}, {name: "GETBIT"},
	{name: "GETDEL"}, {name: "GETEX"}, {name: "GETRANGE"}, {name: "GETSET"},
	{name: "HDEL"}, {name: "HEXISTS"}, {name: "HGET"}, {name: "HGETALL"},
	{name: "HINCRBY"}, {name: "HINCRBYFLOAT"}, {name: "HKEYS"}, {name: "HLEN"},
	{name: "HMGET"}, {name: "HMSET"}, {name: "HRANDFIELD"}, {name: "HSCAN"},
	{name: "HSET"}, {name: "HSETNX"}, {name: "HSTRLEN"}, {name: "HVALS"},
	{name: "INCR"}, {name: "INCRBY"}, {name: "INCRBYFLOAT"}, {name: "INFO"},
	{name: "KEYS"}, {name: "LASTSAVE"}, {name: "LATENCY"}, {name: "LINDEX"},
	{name: "LINSERT"}, {name: "LLEN"}, {name: "LMOVE"}, {name: "LMPOP"},
	{name: "LPOP"}, {name: "LPOS"}, {name: "LPUSH"}, {name: "LPUSHX"},
	{name: "LRANGE"}, {name: "LREM"}, {name: "LSET"}, {name: "LTRIM"},
	{name: "MGET"}, {name: "MIGRATE"}, {name: "MONITOR"}, {name: "MOVE"},
	{name: "MSET"}, {name: "MSETNX"}, {name: "MULTI"}, {name: "OBJECT"},
	{name: "PERSIST"}, {name: "PEXPIRE"}, {name: "PEXPIREAT"}, {name: "PFADD"},
	{name: "PFCOUNT"}, {name: "PFMERGE"}, {name: "PING"}, {name: "PSETEX"},
	{name: "PSUBSCRIBE"}, {name: "PTTL"}, {name: "PUBLISH"},
	{name: "PUBSUB"}, {name: "QUIT"}, {name: "RANDOMKEY"}, {name: "RENAME"},
	{name: "RENAMENX"}, {name: "RESTORE"}, {name: "RPOP"}, {name: "RPOPLPUSH"},
	{name: "RPUSH"}, {name: "RPUSHX"}, {name: "SADD"}, {name: "SAVE"},
	{name: "SCAN"}, {name: "SCARD"}, {name: "SCRIPT"}, {name: "SDIFF"},
	{name: "SDIFFSTORE"}, {name: "SELECT"}, {name: "SET"}, {name: "SETBIT"},
	{name: "SETEX"}, {name: "SETNX"}, {name: "SETRANGE"}, {name: "SHUTDOWN"},
	{name: "SINTER"}, {name: "SINTERCARD"}, {name: "SINTERSTORE"}, {name: "SISMEMBER"},
	{name: "SLAVEOF"}, {name: "SLOWLOG"}, {name: "SMEMBERS"}, {name: "SMISMEMBER"},
	{name: "SMOVE"}, {name: "SORT"}, {name: "SPOP"}, {name: "SRANDMEMBER"},
	{name: "SREM"}, {name: "SSCAN"}, {name: "STRLEN"}, {name: "SUBSCRIBE"},
	{name: "SUNION"}, {name: "SUNIONSTORE"}, {name: "SWAPDB"},
	{name: "SYNC"}, {name: "TIME"}, {name: "TOUCH"}, {name: "TTL"},
	{name: "TYPE"}, {name: "UNLINK"}, {name: "UNSUBSCRIBE"}, {name: "UNWATCH"},
	{name: "WATCH"}, {name: "XACK"}, {name: "XADD"}, {name: "XAUTOCLAIM"},
	{name: "XCLAIM"}, {name: "XDEL"}, {name: "XGROUP"}, {name: "XINFO"},
	{name: "XLEN"}, {name: "XPENDING"}, {name: "XRANGE"}, {name: "XREAD"},
	{name: "XREADGROUP"}, {name: "XREVRANGE"}, {name: "XSETID"}, {name: "XTRIM"},
	{name: "ZADD"}, {name: "ZCARD"}, {name: "ZCOUNT"}, {name: "ZDIFF"},
	{name: "ZDIFFSTORE"}, {name: "ZINCRBY"}, {name: "ZINTER"}, {name: "ZINTERSTORE"},
	{name: "ZLEXCOUNT"}, {name: "ZMPOP"}, {name: "ZPOPMAX"}, {name: "ZPOPMIN"},
	{name: "ZRANDMEMBER"}, {name: "ZRANGE"}, {name: "ZRANGEBYLEX"}, {name: "ZRANGEBYSCORE"},
	{name: "ZRANK"}, {name: "ZREM"}, {name: "ZREMRANGEBYLEX"}, {name: "ZREMRANGEBYRANK"},
	{name: "ZREMRANGEBYSCORE"}, {name: "ZREVRANGE"}, {name: "ZREVRANGEBYSCORE"},
	{name: "ZREVRANK"}, {name: "ZSCAN"}, {name: "ZSCORE"}, {name: "ZUNION"},
	{name: "ZUNIONSTORE"},
}

// BuildGraph makes a synthetic graph where every command is a node inside a
// single synthetic package, with one synthetic "file" per command so the
// runtime mapper can attribute observed heat back to a specific node.
func BuildGraph() *model.StaticGraph {
	graph := &model.StaticGraph{
		Nodes:    make(map[model.NodeID]*model.StaticNode),
		Files:    make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	}

	rootPkg := model.PackageID(1)
	graph.Packages[rootPkg] = &model.Package{Path: ModulePath, Name: moduleName}

	sorted := make([]string, 0, len(commands))
	for _, c := range commands {
		sorted = append(sorted, c.name)
	}
	sort.Strings(sorted)

	rootID := model.NodeID(1)
	rootFileID := model.FileID(1)
	graph.Root = rootID
	graph.Nodes[rootID] = &model.StaticNode{
		Name:   "redis-server",
		ID:     rootID,
		Pkg:    rootPkg,
		Syntax: model.Syntax{Kind: model.SyntaxFuncDecl, File: rootFileID, Start: model.Position{Line: 1}, End: model.Position{Line: 1}},
	}
	graph.Files[rootFileID] = &model.StaticFile{
		ID:        rootFileID,
		Path:      ModulePath + "/redis-server",
		Package:   rootPkg,
		Functions: []model.NodeID{rootID},
	}
	for index, name := range sorted {
		id := model.NodeID(index + 2)
		fileID := model.FileID(index + 2)
		graph.Nodes[id] = &model.StaticNode{
			Name:   name,
			ID:     id,
			Pkg:    rootPkg,
			Syntax: model.Syntax{Kind: model.SyntaxFuncDecl, File: fileID, Start: model.Position{Line: 1}, End: model.Position{Line: 1}},
			In:     []model.NodeID{rootID},
		}
		graph.Nodes[rootID].Out = append(graph.Nodes[rootID].Out, id)
		graph.Files[fileID] = &model.StaticFile{
			ID:        fileID,
			Path:      ModulePath + "/" + name,
			Package:   rootPkg,
			Functions: []model.NodeID{id},
		}
	}

	return graph
}

// CommandNames returns the sorted, de-duplicated list of known command names.
func CommandNames() []string {
	return commandSet().names
}

// IsKnownCommand reports whether name is a well-known Redis command.
func IsKnownCommand(name string) bool {
	_, ok := commandSet().set[name]
	return ok
}

type commandIndex struct {
	names []string
	set   map[string]struct{}
}

var (
	commandIndexOnce sync.Once
	commandIndexVal  commandIndex
)

func commandSet() *commandIndex {
	commandIndexOnce.Do(func() {
		seen := make(map[string]struct{}, len(commands))
		for _, c := range commands {
			seen[c.name] = struct{}{}
		}
		names := make([]string, 0, len(seen))
		for name := range seen {
			names = append(names, name)
		}
		sort.Strings(names)
		commandIndexVal = commandIndex{names: names, set: seen}
	})
	return &commandIndexVal
}
