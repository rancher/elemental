# Scheduling rules for the CI tests

We try to spread the tests as best as we can to avoid SPOT issue as well as not overload our public cloud zone.

| Test type | Day(s) | Hour |
|:---:|:---:|:---:|
| CLI K3s | Monday to Saturday | 6am |
| CLI K3s Upgrade | Monday to Saturday | 8am |
| CLI RKE2 | Monday to Saturday | 7am |
| CLI RKE2 Upgrade | Monday to Saturday | 9am |
| CLI K3s Airgap | Sunday | 2am |
| CLI K3s Scalability | Sunday | 4am |
| CLI Multicluster | Sunday | 5am |
| UI K3s | Monday to Saturday | 2am |
| UI K3s Upgrade | Monday to Saturday | 4am |
| UI RKE2 | Monday to Saturday | 3am |
| UI RKE2 Upgrade | Monday to Saturday | 5am |
| Update tests description | All days | 11pm |

**NOTE:** please note that the GitHub scheduler uses UTC and our GCP runners are deployed in `us-central1`, so UTC-5.
