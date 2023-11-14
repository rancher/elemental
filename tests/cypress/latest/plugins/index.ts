/*
Copyright © 2022 - 2023 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/// <reference types="cypress" />
// eslint-disable-next-line @typescript-eslint/no-var-requires
require('dotenv').config();

/**
 * @type {Cypress.PluginConfig}
 */
// eslint-disable-next-line no-unused-vars
module.exports = (on: Cypress.PluginEvents, config: Cypress.PluginConfigOptions) => {
  // `on` is used to hook into various events Cypress emits
  // `config` is the resolved Cypress config
  const url = process.env.RANCHER_URL || 'https://localhost:8005';
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  const { isFileExist, findFiles } = require('cy-verify-downloads');
  on('task', { isFileExist, findFiles })

  config.baseUrl                  = url.replace(/\/$/, );
  config.env.cache_session        = process.env.CACHE_SESSION || false;
  config.env.chartmuseum_repo     = process.env.CHARTMUSEUM_REPO;
  config.env.cluster              = process.env.CLUSTER_NAME;
  config.env.cypress_tags         = process.env.CYPRESS_TAGS;
  config.env.elemental_ui_version = process.env.ELEMENTAL_UI_VERSION;
  config.env.k8s_version          = process.env.K8S_VERSION_TO_PROVISION;
  config.env.operator_repo        = process.env.OPERATOR_REPO;
  config.env.password             = process.env.RANCHER_PASSWORD;
  config.env.proxy_ip             = process.env.PROXY_IP;
  config.env.proxy                = process.env.PROXY;
  config.env.rancher_version      = process.env.RANCHER_VERSION;
  config.env.ui_account           = process.env.UI_ACCOUNT;
  config.env.upgrade_image        = process.env.UPGRADE_IMAGE;
  config.env.upgrade_os_channel   = process.env.UPGRADE_OS_CHANNEL;
  config.env.username             = process.env.RANCHER_USER;

  return config;
};
