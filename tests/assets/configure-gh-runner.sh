#!/bin/bash

# !! LET THIS SCRIPT IN THE GITHUB REPOSITORY !!
# !!  EVEN IF IT IS NOT USED DIRECTLY BY CI   !!
# !!    IT IS USED TO BOOSTRAP THE RUNNER     !!

# Variable(s)
SSD=/dev/nvme0n1
GH_USER=gh-runner
HOST_PREFIX="elemental-ci-"
HOST_PATTERN="^${HOST_PREFIX}[a-f0-9][a-f0-9]*-[a-f0-9][a-f0-9]*-[a-f0-9][a-f0-9]*-[a-f0-9][a-f0-9]*-[a-f0-9][a-f0-9]*\$"
UNVALID_HOSTNAME=1
GCLOUD_BIN=/opt/google-cloud-sdk/bin/gcloud

# Just for logs
echo "$0: started"

# Hostname can take times to be set
(( SECONDS_TO_WAIT = SECONDS + 60 ))
while (( SECONDS < SECONDS_TO_WAIT )); do
  HOSTNAME=$(hostname)
  echo "$0: checking hostname: ${HOSTNAME}"

  # Break the loop if hostname matches the pattern
  if [[ "${HOSTNAME}" =~ ${HOST_PATTERN} ]]; then
    UNVALID_HOSTNAME=0
    break
  fi

  # Wait a little before checking again
  sleep 10
done

# Continue only if hostname matches the pattern
if (( UNVALID_HOSTNAME )); then
  # End script without error
  echo "$0: wrong hostname detected, not a runner! stopped"
  exit 0
fi

# Extract UUID
# NOTE: HOSTNAME needs to be set here, to be sure to have the correct value
HOSTNAME=$(hostname)
UUID=${HOSTNAME#${HOST_PREFIX}}

# Get PAT token
PAT_TOKEN=$(${GCLOUD_BIN} secrets versions access latest --secret="PAT_TOKEN_${UUID}")
if [[ -z "${PAT_TOKEN}" ]]; then
  echo "$0: PAT token not found! stopped"
  exit 1
fi

# Configure Local SSD
mkfs -t xfs -f ${SSD}

# Configure gh-runner account
mkdir -p /home/${GH_USER}
mount ${SSD} /home/${GH_USER}
useradd -d /home/${GH_USER} -g users -G docker,libvirt,google-sudoers -M ${GH_USER}
chown -R ${GH_USER}:users /home/${GH_USER}

# Install and configure GH runner (should be run with 'gh-runner' users)

## Generate registration token
GH_REPO=rancher/elemental
TOKEN=$(curl \
          -X POST \
          -H "Accept: application/vnd.github+json" \
          -H "Authorization: Bearer ${PAT_TOKEN}" \
          https://api.github.com/repos/${GH_REPO}/actions/runners/registration-token | jq -r '.token')

# Generate gh-runner script
GH_SCRIPT=/home/${GH_USER}/${0##*/}
cat > ${GH_SCRIPT} <<EOF
#!/bin/bash

# Variable(s)
TAR_FILE=runner.tar.gz

# Create a folder
cd /home/${GH_USER}
mkdir -p actions-runner && cd actions-runner

# Get the latest runner version
(( SECONDS_TO_WAIT = SECONDS + 30 ))
while (( SECONDS < SECONDS_TO_WAIT )); do
  PKG=\$(wget -q -O - https://api.github.com/repos/actions/runner/releases/latest \\
        | awk '/browser_download_url.*\/actions-runner-linux-x64-.*[0-9].tar.gz/ { print \$2 }')

  # PKG should contains '.tar.gz' string
  [[ \"\${PKG}\" =~ .tar.gz ]] && break

  # Wait a little before trying again
  sleep 5
done

# Download the latest runner package
(( SECONDS_TO_WAIT = SECONDS + 90 ))
while (( SECONDS < SECONDS_TO_WAIT )); do
  wget -q -O \${TAR_FILE} \${PKG//\\"/}

  # If file is here with a size > 0 then it should be OK
  [[ -s \${TAR_FILE} ]] && break

  # Wait a little before trying again
  sleep 5
done

# Extract the installer
tar xzf \${TAR_FILE}

# Create the runner and start the configuration experience
./config.sh \\
  --ephemeral \\
  --unattended \\
  --url https://github.com/${GH_REPO} \\
  --token ${TOKEN} \\
  --labels ${UUID} \\
  --name ${HOSTNAME}

# Last step, run it! (in root)
sudo ./svc.sh install
sudo ./svc.sh start
EOF

# Execute script
sudo -u ${GH_USER} bash ${GH_SCRIPT}

# End script without error
echo "$0: stopped"
exit 0
