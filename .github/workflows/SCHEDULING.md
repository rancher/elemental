# Scheduling rules for the CI tests

We try to spread the tests as best as we can to avoid SPOT issue as well as not overload our public cloud zone.

| Test type | Day(s) | Hour | Zones |
|:---:|:---:|:---:|:---:|
| CLI K3s | Monday to Saturday | 3am | us-central1-c |
| CLI K3s Upgrade | Monday to Saturday | 5am | us-central1-c |
| CLI RKE2 | Monday to Saturday | 3am | us-central1-f |
| CLI RKE2 Upgrade | Monday to Saturday | 5am | us-central1-f |
| CLI K3s Airgap | Sunday | 1am | us-central1-c |
| CLI K3s Scalability | Sunday | 2am | us-central1-f |
| CLI Multicluster | Sunday | 5am | us-central1-b |
| CLI Rancher Manager Devel | Sunday | 8am | us-central1-c |
| UI Rancher Manager Devel | Sunday | 12am | us-central1-a |
| UI K3s | Monday to Saturday | 2am | us-central1-a |
| UI K3s Upgrade | Monday to Saturday | 4am | us-central1-a |
| UI RKE2 | Monday to Saturday | 2am | us-central1-b |
| UI RKE2 Upgrade | Monday to Saturday | 4am | us-central1-b |
| Update tests description | All days | 11pm | us-central1 |

**NOTE:** please note that the GitHub scheduler uses UTC and our GCP runners are deployed in `us-central1`, so UTC-5.
