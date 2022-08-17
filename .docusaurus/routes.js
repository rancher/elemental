import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/__docusaurus/debug',
    component: ComponentCreator('/__docusaurus/debug', 'd33'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/config',
    component: ComponentCreator('/__docusaurus/debug/config', 'ca0'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/content',
    component: ComponentCreator('/__docusaurus/debug/content', '3e7'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/globalData',
    component: ComponentCreator('/__docusaurus/debug/globalData', '3e6'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/metadata',
    component: ComponentCreator('/__docusaurus/debug/metadata', '64f'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/registry',
    component: ComponentCreator('/__docusaurus/debug/registry', '231'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/routes',
    component: ComponentCreator('/__docusaurus/debug/routes', '2c8'),
    exact: true
  },
  {
    path: '/blog',
    component: ComponentCreator('/blog', '869'),
    exact: true
  },
  {
    path: '/blog/archive',
    component: ComponentCreator('/blog/archive', '162'),
    exact: true
  },
  {
    path: '/blog/first-blog-post',
    component: ComponentCreator('/blog/first-blog-post', '0a7'),
    exact: true
  },
  {
    path: '/blog/long-blog-post',
    component: ComponentCreator('/blog/long-blog-post', 'eee'),
    exact: true
  },
  {
    path: '/blog/mdx-blog-post',
    component: ComponentCreator('/blog/mdx-blog-post', '77d'),
    exact: true
  },
  {
    path: '/blog/tags',
    component: ComponentCreator('/blog/tags', '9c4'),
    exact: true
  },
  {
    path: '/blog/tags/docusaurus',
    component: ComponentCreator('/blog/tags/docusaurus', 'd13'),
    exact: true
  },
  {
    path: '/blog/tags/facebook',
    component: ComponentCreator('/blog/tags/facebook', 'cf2'),
    exact: true
  },
  {
    path: '/blog/tags/hello',
    component: ComponentCreator('/blog/tags/hello', 'adc'),
    exact: true
  },
  {
    path: '/blog/tags/hola',
    component: ComponentCreator('/blog/tags/hola', '3d5'),
    exact: true
  },
  {
    path: '/blog/welcome',
    component: ComponentCreator('/blog/welcome', '3ae'),
    exact: true
  },
  {
    path: '/',
    component: ComponentCreator('/', '655'),
    routes: [
      {
        path: '/',
        component: ComponentCreator('/', '7d9'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/architecture',
        component: ComponentCreator('/architecture', 'a6e'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/cloud-config-reference',
        component: ComponentCreator('/cloud-config-reference', '124'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/customizing',
        component: ComponentCreator('/customizing', '93d'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/elemental-plans',
        component: ComponentCreator('/elemental-plans', 'bdd'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/installation',
        component: ComponentCreator('/installation', '2e7'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/kubernetesversions',
        component: ComponentCreator('/kubernetesversions', '73b'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/machineregistration-reference',
        component: ComponentCreator('/machineregistration-reference', '384'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/quickstart',
        component: ComponentCreator('/quickstart', '99e'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/upgrade',
        component: ComponentCreator('/upgrade', 'a11'),
        exact: true,
        sidebar: "tutorialSidebar"
      }
    ]
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
