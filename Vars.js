import React from 'react';

export default function Vars({children, props}) {
  return (
    <span>{children}</span>
  )
}
// export default function vars() {
//   return (
//     {
//       slackName: '#elemental'
//     }
//   )
// /*   slackName: function slackName() {
//     return (
//       '#elemental'
//     )
//   },

//   slackUrl: function slackUrl() {
//     return (
//       'https://rancher-users.slack.com/channels/elemental'
//     )
//   } */
// }


/* export const elemental_slack_name = "#elemental"
export const elemental_slack_url = "https://rancher-users.slack.com/channels/elemental";
export const elemental_toolkit_name = "Elemental Toolkit";
export const elemental_toolkit_url = "https://rancher.github.io/elemental-toolkit";
export const elemental_operator_name = "Elemental Operator";
export const elemental_operator_url = "https://github.com/rancher/elemental-operator";
export const elemental_cli_name = "Elemental CLI";
export const elemental_cli_url = "https://github.com/rancher/elemental-cli";
export const ranchersystemagent_name = "Rancher System Agent";
export const ranchersystemagent_url = "https://github.com/rancher/system-agent"; */