// Check the Cypress tags
// Implemented but not used yet
export const isCypressTag = (tag: string) => {
  return (new RegExp(tag)).test(Cypress.env("cypress_tags"));
}

// Check the K8s version
export const isK8sVersion = (version: string) => {
  version = version.toLowerCase();
  return (new RegExp(version)).test(Cypress.env("k8s_version"));
}

// Check the Elemental operator version
export const isOperatorVersion = (version: string) => {
  return (new RegExp(version)).test(Cypress.env("operator_repo"));
}

// Check rancher manager version
export const isRancherManagerVersion = (version: string) => {
  return (new RegExp(version)).test(Cypress.env("rancher_version"));
}

// Check Elemental UI version
export const isUIVersion = (version: string) => {
  return (new RegExp(version)).test(Cypress.env("elemental_ui_version"));
}
