# GoSaaS

An Oauth2 authentication server cli for building software as a service.
GoSaaS creates sets up the following to jumpstart your SaaS project:

* GoSaaS provides a complete OAuth2 login flow for protected router paths. Data is saved to Auth0's database as well as in sessions
* GoSaaS saves user data to sessions, which is retrieved and sent to backend apis as a protobuf message. This data is also used to render templates.
* GoSaaS saves user tokens to sessions, which are retrieved and sent to backend apis as protobuf messages.
* GoSaaS uses Sprig's Funcmap for rendering templates ref: github.com/Masterminds/sprig


## RoadMap

- [ ]  Protected endpoints by Stripe subscription



## Download

    go get github.com/autom8ter/gosaas
    
## Usage


**Command:**

    gosaas
    
**Output:**
```text

---------------------------------------------------
   .aMMMMP .aMMMb  .dMMMb  .aMMMb  .aMMMb  .dMMMb
  dMP"    dMP"dMP dMP" VP dMP"dMP dMP"dMP dMP" VP
 dMP MMP"dMP dMP  VMMMb  dMMMMMP dMMMMMP  VMMMb  
dMP.dMP dMP.aMP dP .dMP dMP dMP dMP dMP dP .dMP  
VMMMP"  VMMMP"  VMMMP" dMP dMP dMP dMP  VMMMP"   
---------------------------------------------------

Usage:
  gosaas [command]

Available Commands:
  config      debug config
  flags       debug flags
  help        Help about any command
  serve       start the GoSaaS server

Flags:
      --config string   config file (default is $HOME/.gosaas.yaml)
  -h, --help            help for gosaas

Use "gosaas [command] --help" for more information about a command.

```

---

**Command:**

    gosaas flags
    
**Output:**

```text
DebugFlags called on config

DebugFlags called on flags
flags
  -, --config []     [L]
  -h, --help [false]  false   [L]

DebugFlags called on help

DebugFlags called on serve
serve
  -a, --addr [:8080]  :8080   [L]
  -, --blog [static/blog.html]  static/blog.html   [L]
  -, --home [static/home.html]  static/home.html   [L]
  -, --loggedin [static/loggedin.html]  static/loggedin.html   [L]

```

---

**Command:**

    gosaas config
    
**Output:**

```text

Defaults:
map[string]interface {}{"addr":":8080", "home":"static/home.html", "loggedin":"static/loggedin.html", "blog":"static/blog.html"}

```

---

**Command:**

    gosaas serve -h
    
**Output:**

```text

start the GoSaaS server

Usage:
  gosaas serve [flags]

Flags:
  -a, --addr string       address to serve on (default ":8080")
      --blog string       path to blog template (default "static/blog.html")
  -h, --help              help for serve
      --home string       path to home template (default "static/home.html")
      --loggedin string   path to loggedin template (default "static/loggedin.html")

Global Flags:
      --config string   config file (default is $HOME/.gosaas.yaml)

```