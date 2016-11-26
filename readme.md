# Cryodns
[![Build Status (Travis)][![Build Status (Travis)](https://travis-ci.org/unixvoid/cryodns.svg?branch=master)](https://travis-ci.org/unixvoid/cryodns)]
This is the DNS service running my personal site [unixvoid](https://unixvoid.com).  
It was built as a drop-in replacement for a popular dynamic DNS updater
[DYN](https://dyn.com).  The motivation for this project was not monitary, I
just wanted a tool to more fit my needs.. and since I make a lot of DNS based
tools I figured that this would feel right at home.  
The main use of this tool is in conjuction with an updater (cryodns updater in
progress now) to update any amount of domains very fast.  The exposed API allows
for easy consumption and updating.

## Running cryodns
There are 3 main ways to run cryodns:  

1. **Docker**: we have cryodns pre-packaged over on the [dockerhub](https://hub.docker.com/r/unixvoid/cryodns/), go grab the latest and run: 
`docker run -d -p 8080:8080 -p 53:53 unixvoid/cryodns`.

2. **ACI/rkt**: we have public rkt images hosted on the site! check them out [here](https://cryo.unixvoid.com/bin/rkt/cryodns/) or go give us a fetch for 64bit machines!
`rkt fetch unixvoid.com/cryodns`.  This image can be run with rkt or you can
grab our handy [service file](https://github.com/unixvoid/cryodns/blob/master/deps/cryodns.service)

3. **From Source**: Are we not compiled for your architecture? Wanna hack on the source?  Lets bulid and deploy:  
  `make dependencies`  
  `make run`  
  If you want to build a docker use: `make docker`  
  If you want to build an ACI use: `make aci`

## API guide
cryodns exposes an api for adding, removing, and querying dns entries.  The
following is the specification for endpoints and their protocols.
- `/dns` : `GET` : endpoint for getting a list of registered dns entries.
  - example: `curl localhost:8080/dns`
- `/dns` : `POST` : endpoint for adding (or updating) dns entries. dnstype will default to 'A'
  - parameters: `domain``sec``dnstype``value`
  - example: `curl localhost:8080/dns`
- `/register` : `GET` : endpoint for registering master admin (returns security
  token). If the instance is not set to `bootstrap` in the config, this will
  register the cryodns instance and return sec key to user.
  - example: `curl localhost:8080/register`
- `/remove` : `` : endpoint for removing dns entries. if `dnstype` is not set,
  it will remove all entries
  - parameters: `domain``sec``dnstype`
  - example: `curl -d domain=unixvoid.com -d sec=<<SEC_TOKEN> localhost:8080/dns`
- `rotate` : `POST` : endpoint to rotate security token. this takes the current
  security token and will generate a new one.
  - parameters: `sec`
  - example: `curl -d sec=<SEC_TOKEN> localhost:8080/rotate`


## Configuration
The configuration is very straightforward, we can take a look at the default
config file and break it down.
```
[cryo]								# this section is the start of the servers main config.
	loglevel		= "debug"		# loglevel, this can be [debug, cluster, info, error]
	dnsport			= 53			# port for DNS listener to bind to
	apiport			= 8080			# port for API listener to bind to
	ttl				= 0				# default TTL for every DNS record
	bootstrap		= true			# bootstrap with a default security token or leave registration up to user. If selected, sec key will be generated on first boot, otherwise the /register api endpoint will be needed for initial registration
	sectokensize	= 25			# length of security token in characters
	secdictionary 	= "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"	# dictionary for security token

[upstream]							# this section is the start of the upstream DNS config
	server			= "8.8.8.8:53"	# upstream dns address AND port.  This will be used if DNS record is not in the database.  No upstream requests are used if this is left blank

[redis]								# this section is the start of redis configurations
	host			= "localhost:6379"# address and port of the redis server to be used
	password		= ""			# password for the redis server
```

## Setting up with your domain name provider
To start off, every domain provider will probably be different.  If the fields
are not identical with the listed ones this is fine, following this outline
should make it pretty simple for any versed individual to setup.  
  
**Precursor**: it is recommended to set up cryodns on a server **before** making
the following changes to your domains DNS settings, this will allow the server
to start taking requests immediately.  
  
1. Find a suitable place to run your DNS nameserver.  A suitible server is one
   with a *static ip*.  I know that a large use of this project is to deal with
   non-static ip's, but every nameserver has to be static to work properly.
   There are many cloud infrastructure providers that allow free accounts as
   long as recource usage is low.  Luckily our server does not take many
   recources (maybe a little network bandwidth).  To get cryodns running, check out the `running
   cryodns` section above.  
2. Add all entries that you will be authoritive for in your cryodns database.
   For this, take a look at the API guide for **adding entries**.  
3. Collect the static-ip of the server running your version of cryodns  
4. Visit your domain name providers website and find the `nameservers`/`advanced DNS` 
   section. You will need to add ANAME entries for your nameserver. One example
   of this is what I use for `unixvoid.com`.  I will add two ANAME entries
   `ns1.unixvoid.com` and `ns2.unixvoid.com`.  I will set these to the static ip 
   that we collected from step 3.  
5. Once we have the settings set for our provider we will need the new ip's to
   propagate, this can take some time, go get yourself a big ol cup of
   **coffee**.  
6. Tail those logs and make sure all requests are working correctly.  
7. Profit.

## How I use cryodns
I run a lot of *different* DNS services.  I prefer to save money and put many
*different* services on one box.  I use a DNS proxy
[dproxy](https://github.com/unixvoid/dproxy) to stand out front on port 53 and
reverse proxy requests at a DNS level back to cryodns (one example of a DNS
server running behind dproxy).  When a request for `unixvoid.com` comes in, the
request is sent to cryodns which fields all dns requests for the domain, here is
what my dproxy config looks like:
```
[*.unixvoid.com]
        address         = 172.31.43.212
        port            = 8054

[unixvoid.com]
        address         = 172.31.43.212
        port            = 8054
```
This takes all requests (domain **and** subdomains) and routes them over to the
rkt container running cryodns.  Please note that dproxy is not required for
this, I have used it in the example for demonsration purposes of my personal
environment.  Running cryodns by itself means much less moving pieces and is
very easy to set up.

### Milestones
- Proper DNS responses for non-autoritive requests.
