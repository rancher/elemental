/// <reference types="cypress" />
require('dotenv').config();

/**
 * @type {Cypress.PluginConfig}
 */
// eslint-disable-next-line no-unused-vars
module.exports = (on: Cypress.PluginEvents, config: Cypress.PluginConfigOptions) => {
  // `on` is used to hook into various events Cypress emits
  // `config` is the resolved Cypress config
  const url = process.env.RANCHER_URL || 'https://localhost:8005';
  const { isFileExist, findFiles } = require('cy-verify-downloads');
  on('task', { isFileExist, findFiles })

  config.baseUrl = url.replace(/\/$/, '');

  config.env.username = process.env.RANCHER_USER;
  config.env.password = process.env.RANCHER_PASSWORD;
  config.env.cluster = process.env.CLUSTER_NAME;
  config.env.cache_session = process.env.CACHE_SESSION || false;
  config.env.k8s_version = process.env.K8S_VERSION_TO_PROVISION;
  config.env.ui_account = process.env.UI_ACCOUNT;
  config.env.operator_version = process.env.OPERATOR_VERSION;
  config.env.proxy = process.env.PROXY;
  config.env.proxy_ip = process.env.PROXY_IP;

  return config;
};
