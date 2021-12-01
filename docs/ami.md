# Amazon AMIs

AMIs for RancherOS are published under the owner ID `275947076441` in the `us-west-1` and `us-west-2` regions
currently.

```bash
aws --region=us-west-1 ec2 describe-images --owners 275947076441
aws --region=us-west-2 ec2 describe-images --owners 275947076441
```
