import { defineConfig } from 'cypress'

const qaseAPIToken = process.env.QASE_API_TOKEN

export default defineConfig({
  defaultCommandTimeout: 10000,
  reporter: 'cypress-qase-reporter',
  reporterOptions: {
    'apiToken': qaseAPIToken,
    'projectCode': 'ELEMENTAL',
    'logging': true,
  },
  e2e: {
    // We've imported your old cypress plugins here.
    // You may want to clean this up later by importing these.
    setupNodeEvents(on, config) {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      return require('./plugins/index.ts')(on, config)
    },
    experimentalSessionAndOrigin: true,
    supportFile: './support/e2e.ts',
    fixturesFolder: './fixtures',
    screenshotsFolder: './screenshots',
    videosFolder: './videos',
    downloadsFolder: './downloads',
    specPattern: 'e2e/unit_tests/*.spec.ts',
  },
})
