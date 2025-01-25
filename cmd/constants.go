package cmd

const (
	semlinkTagXattrKey  = "user.semlink.tags"
	semlinkTypeXattrKey = "user.semlink.type"
	registryDir         = ".semlink"
	databaseDir         = ".config/semlink"
	databaseFile        = "semlink.sqlite"
	registryFile        = "registry.json"
	defaultType         = "source"
	registryPermissions = 0755

	RECEIVER = "receiver"
	VIRTUAL  = "virtual"
	SOURCE   = "source"
)
