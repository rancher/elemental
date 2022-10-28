import React from 'react';

// Global variables to be adapted
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

/***
 * DO NOT TOUCH -- Docusaurus component logic
***/
export default function Vars({children, name, link}) {
  // Check if the link variable is set
  if (link !== undefined) {
    // Check if the link has a trailing path
    const linkparts = link.split(/\/(.*)/s)
    // Sets the URL to the global variable
    var linkurl = `${data[linkparts[0]]}`
    // Adds the trailing path if it exists
    if (linkparts.length > 2) {
      linkurl = `${data[linkparts[0]]}/${linkparts[1]}`
    }
    if (children !== undefined) {
      return (
        <a href={linkurl}>{children}</a>
      )
    }
    else {
      return (
        <a href={linkurl}>{data[name]}</a>
      )
    }
  }
  else {
    return (
      <span>{data[name]}</span>
    )
  }
}
