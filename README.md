# Usage
`env2conf`: prints the environment as json

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
