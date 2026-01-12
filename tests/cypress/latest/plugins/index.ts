/*
Copyright © 2022 - 2026 SUSE LLC

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

  config.baseUrl = url.replace(/\/$/,);
  config.env.arch = process.env.ARCH;
  config.env.boot_type = process.env.BOOT_TYPE;
  config.env.cache_session = process.env.CACHE_SESSION || false;
  config.env.chartmuseum_repo = process.env.CHARTMUSEUM_REPO;
  config.env.cluster = process.env.CLUSTER_NAME;
  config.env.container_stable_os_version = process.env.CONTAINER_STABLE_OS_VERSION;
  config.env.cypress_tags = process.env.CYPRESS_TAGS;
  config.env.elemental_dev_version = process.env.ELEMENTAL_DEV_VERSION;
  config.env.elemental_ui_version = process.env.ELEMENTAL_UI_VERSION;
  config.env.k8s_downstream_version = process.env.K8S_DOWNSTREAM_VERSION;
  config.env.operator_install_type = process.env.OPERATOR_INSTALL_TYPE;
  config.env.operator_repo = process.env.OPERATOR_REPO;
  config.env.os_version_install = process.env.OS_VERSION_INSTALL;
  config.env.os_version_target = process.env.OS_VERSION_TARGET;
  config.env.os_version_to_test = process.env.OS_VERSION_TO_TEST;
  config.env.password = process.env.RANCHER_PASSWORD;
  config.env.proxy_ip = process.env.PROXY_IP;
  config.env.proxy = process.env.PROXY;
  config.env.rancher_git_chart = process.env.RANCHER_GIT_CHART;
  config.env.rancher_channel = process.env.RANCHER_CHANNEL;
  config.env.rancher_version = process.env.RANCHER_VERSION;
  config.env.iso_stable_os_version = process.env.ISO_STABLE_OS_VERSION;
  config.env.upgrade_from_version = process.env.UPGRADE_FROM_VERSION;
  config.env.upgrade_image = process.env.UPGRADE_IMAGE;
  config.env.upgrade_os_channel = process.env.UPGRADE_OS_CHANNEL;
  config.env.username = process.env.RANCHER_USER;

  return config;
};
