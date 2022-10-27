import React from 'react';

export const data = {
  elemental_slack_name: "#elemental",
  elemental_slack_url: "https://rancher-users.slack.com/channels/elemental",
  elemental_toolkit_name: "Elemental Toolkit",
  elemental_toolkit_url: "https://rancher.github.io/elemental-toolkit",
  elemental_operator_name: "Elemental Operator",
  elemental_operator_url: "https://github.com/rancher/elemental-operator",
  elemental_cli_name: "Elemental CLI",
  elemental_cli_url: "https://github.com/rancher/elemental-cli",
  elemental_iso_name: "Elemental ISO",
  elemental_register_name: "Elemental Register client",
  ranchersystemagent_name: "Rancher System Agent",
  ranchersystemagent_url: "https://github.com/rancher/system-agent",
}

export default function Vars({children, name, link}) {

  if (link !== undefined) {
    if (children !== undefined) {
      return (
        <a href={data[link]}>{children}</a>
      )
    }
    else {
      return (
        <a href={data[link]}>{data[name]}</a>
      )
    }
  }
  else {
    return (
      <span>{data[name]}</span>
    )
  }
}
