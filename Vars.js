import React from 'react';

export default function Vars({children, name}) {
  const data = {
    elemental_slack_name: "#elemental",
    elemental_slack_url: <a href="https://rancher-users.slack.com/channels/elemental">{children}</a>,
    elemental_toolkit_name: "Elemental Toolkit",
    elemental_toolkit_url: <a href="https://rancher.github.io/elemental-toolkit">{children}</a>,
    elemental_operator_name: "Elemental Operator",
    elemental_operator_url: <a href="https://github.com/rancher/elemental-operator">{children}</a>,
    elemental_cli_name: "Elemental CLI",
    elemental_cli_url: <a href="https://github.com/rancher/elemental-cli">{children}</a>,
    ranchersystemagent_name: "Rancher System Agent",
    ranchersystemagent_url: <a href="https://github.com/rancher/system-agent">{children}</a>,
  }

  return (
    <span>{data[name]}</span>
  )
}
