## Usage
`env2conf`: prints the environment as json, xml, toml, or yaml

`env2conf prefix_`: exclude environment variables that don't start with "prefix_"

`env2conf prefix.`: exclude environment variables that don't start with "prefix."

```
env - \
	'conf.servers[0].ip=192.168.1.1' \
	'conf.servers[1].ip=192.0.2.42' \
	'conf.servers[0].name="database server"' \
	'conf.maxClients=7' \
	'conf.tickTackToe[2][2]=X' \
	env2conf conf.
```
```
{
  "maxClients": 7,
  "servers": [
    {
      "ip": "192.168.1.1",
      "name": "database server"
    },
    {
      "ip": "192.0.2.42"
    }
  ],
  "tickTackToe": [
    null,
    null,
    [
      null,
      null,
      "X"
    ]
  ]
}
```

## Options
```
  --fmt string
        Output format. Can be json, xml, toml, yaml (default "json")
  --pid int
        Read environment from given PID (defaults to self). This can be usefull if your shell strips environment variables containing special charicters.
  --root string
        Root tag to be used when generating XML (default "config")
  --underscores
        Use _ as a field seperator. Use __ (two underscores) for a literal _
```
