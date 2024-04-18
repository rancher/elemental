# Scheduling rules for the CI tests

We try to spread the tests as best as we can to avoid SPOT issue as well as not overload our public cloud zone.

| Test type | Day(s) | Hour |
|:---:|:---:|:---:|
| CLI K3s | Monday to Saturday | 3am |
| CLI K3s Upgrade | Monday to Saturday | 7am |
| CLI RKE2 | Monday to Saturday | 5am |
| CLI RKE2 Upgrade | Monday to Saturday | 8am |
| CLI K3s Airgap | Sunday | 1am |
| CLI K3s Scalability | Sunday | 2am |
| CLI Multicluster | Sunday | 5am |
| CLI Rancher Manager Devel | Sunday | 8am |
| UI Rancher Manager Devel | Sunday | 12am |
| UI K3s | Monday to Saturday | 11pm |
| UI K3s Upgrade | Monday to Saturday | 1am |
| UI RKE2 | Monday to Saturday | 0am |
| UI RKE2 Upgrade | Monday to Saturday | 2am |
| Update tests description | All days | 11pm |

**NOTE:** please note that the GitHub scheduler uses UTC and our GCP runners are deployed in `us-central1`, so UTC-5.
