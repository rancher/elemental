# Open Build Service integration files

This folder includes specific files to integrate with the [Open Build Service](https://build.opensuse.org) (OBS) by
[openSUSE](https://www.opensuse.org). Includes the workflows definition file (`workflows.yml`) that defines
actions for pull request, push tag and commit events.

In addition it also includes any required spec file or Dockerfile, usually these
are adapted to OBS and can't be used outside the OBS context.
