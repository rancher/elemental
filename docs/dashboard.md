# Dashboard/UI

The Rancher UI is running by default on port `:8443`.  There default
`admin` user password set is a long random string.  You can run `rancherd reset-admin` to
get a new `admin` password to login.

To disable the Rancher UI from running on a host port, or to change the
default hostPort used the below configuration.

```yaml
#cloud-config
rancherd:
  rancherValues:
    # Setting the host port to 0 will disable the hostPort, default is 8443
    hostPort: 0
```
