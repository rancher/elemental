# Scheduling rules for the CI tests

We try to spread the tests as best as we can to avoid SPOT issue as well as not overload our public cloud region.

Tests that are scheduled (can be manually triggered as well):
| Test type | Day(s) | Hour | Region |
|:---:|:---:|:---:|:---:|
| CLI RKE2 | Monday | 2am | europe-west1 |
| CLI Full backup/restore (migration) | Wednesday | 2am | europe-west1 |
| UI K3s | Tuesday | 2am | europe-west1 |
| Update tests description | All days | 11pm | europe-west1 |

Tests that are not scheduled (but can be manually triggered):
| Test type | Zones |
|:---:|:---:|
| CLI K3s | europe-west1 |
| CLI K3s Airgap | europe-west1 |
| CLI K3s Downgrade | europe-west1 |
| CLI K3s IBS Stable | europe-west1 |
| CLI K3s OBS Dev | europe-west1 |
| CLI K3s OBS Staging | europe-west1 |
| CLI K3s Scalability | europe-west1 |
| CLI K3s SELinux | europe-west1 |
| CLI K3s Upgrade | europe-west1 |
| CLI Multicluster | europe-west1 |
| CLI Regression | europe-west1 |
| CLI RKE2 Upgrade | europe-west1 |
| CLI RKE2 IBS Stable | europe-west1 |
| CLI RKE2 OBS Dev | europe-west1 |
| CLI RKE2 OBS Staging | europe-west1 |
| CLI OBS Manual | europe-west1 |
| CLI OBS Manual Upgrade | europe-west1 |
| UI K3s IBS Stable | europe-west1 |
| UI K3s OBS Dev | europe-west1 |
| UI K3s OBS Staging | europe-west1 |
| UI K3s Upgrade | europe-west1 |
| UI Markeplace | europe-west1 |
| UI Markeplace Upgrade| europe-west1 |
| UI RKE2 | europe-west1 |
| UI RKE2 IBS Stable | europe-west1 |
| UI RKE2 OBS Dev | europe-west1 |
| UI RKE2 OBS Staging | europe-west1 |
| UI RKE2 Upgrade | europe-west1 |
| UI OBS Manual | europe-west1 |
| UI OBS Manual Upgrade | europe-west1 |

**NOTE:** please note that the GitHub scheduler uses UTC and our GCP runners are deployed in `europe-west1`
