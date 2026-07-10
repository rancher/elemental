# Scheduling rules for the CI tests

We try to spread the tests as best as we can to avoid SPOT issue as well as not overload our public cloud zone.

Tests that are scheduled (can be manually triggered as well):
| Test type | Day(s) | Hour | Zones |
|:---:|:---:|:---:|:---:|
| CLI RKE2 | Monday | 4am | us-central1-c |
| CLI Full backup/restore (migration) | Wednesday | 4am | us-central1-a |
| UI K3s | Tuesday | 4am | us-central1-a |
| Update tests description | All days | 11pm | us-central1 |

Tests that are not scheduled (but can be manually triggered):
| Test type | Zones |
|:---:|:---:|
| CLI K3s | us-central1-a |
| CLI K3s Airgap | us-central1-b |
| CLI K3s Downgrade | us-central1-c |
| CLI K3s IBS Stable | us-central1-f |
| CLI K3s OBS Dev | us-central1-f |
| CLI K3s OBS Staging | us-central1-f |
| CLI K3s Scalability | us-central1-f |
| CLI K3s SELinux | us-central1-a |
| CLI K3s Upgrade | us-central1-b |
| CLI Multicluster | us-central1-f |
| CLI Regression | us-central1-f |
| CLI RKE2 Upgrade | us-central1-f |
| CLI RKE2 IBS Stable | us-central1-f |
| CLI RKE2 OBS Dev | us-central1-f |
| CLI RKE2 OBS Staging | us-central1-f |
| CLI OBS Manual | us-central1-f |
| CLI OBS Manual Upgrade | us-central1-f |
| UI K3s IBS Stable | us-central1-f |
| UI K3s OBS Dev | us-central1-f |
| UI K3s OBS Staging | us-central1-f |
| UI K3s Upgrade | us-central1-b |
| UI Markeplace | us-central1-f |
| UI Markeplace Upgrade| us-central1-f |
| UI RKE2 | us-central1-c |
| UI RKE2 IBS Stable | us-central1-f |
| UI RKE2 OBS Dev | us-central1-f |
| UI RKE2 OBS Staging | us-central1-f |
| UI RKE2 Upgrade | us-central1-f |
| UI OBS Manual | us-central1-f |
| UI OBS Manual Upgrade | us-central1-f |

**NOTE:** please note that the GitHub scheduler uses UTC and our GCP runners are deployed in `us-central1`, so UTC-5.
