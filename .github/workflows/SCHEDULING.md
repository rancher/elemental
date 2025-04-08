# Scheduling rules for the CI tests

We try to spread the tests as best as we can to avoid SPOT issue as well as not overload our public cloud zone.

| Test type | Day(s) | Hour | Zones |
|:---:|:---:|:---:|:---:|
| CLI K3s | Monday/Wednesday | 2am | us-central1-c |
| CLI K3s Upgrade | Monday/Wednesday | 4am | us-central1-c |
| CLI RKE2 | Monday/Wednesday | 6am | us-central1-f |
| CLI RKE2 Upgrade | Monday/Wednesday | 8am | us-central1-f |
| CLI K3s Airgap | Friday | 4am | us-central1-c |
| CLI K3s Scalability | Not scheduled anymore | X | us-central1-f |
| CLI K3s SELinux | Not scheduled anymore | X | us-central1-c |
| CLI Multicluster | Not scheduled anymore | X | us-central1-b |
| CLI Regression | Friday | 8am | us-central1-c |
| CLI K3s Downgrade | Friday | 6am | us-central1-b |
| CLI Full backup/restore (migration) | Friday | 2am | us-central1-c |
| UI K3s | Tuesday/Thursday | 2am | us-central1-a |
| UI K3s Upgrade | Tuesday/Thursday | 4am | us-central1-a |
| UI RKE2 | Tuesday/Thursday | 6am | us-central1-b |
| UI RKE2 Upgrade | Tuesday/Thursday | 8am | us-central1-b |
| Update tests description | All days | 11pm | us-central1 |

**NOTE:** please note that the GitHub scheduler uses UTC and our GCP runners are deployed in `us-central1`, so UTC-5.
