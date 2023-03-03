// Allow to filter Cypress tests with tags
const TestFilters = (givenTags, runTest) => {
  const tags = Cypress.env('cypress_tags').split(',')
  const isFound = givenTags.some((givenTag) => tags.includes(givenTag))

    if (isFound) {
      runTest()
    }
};

export default TestFilters
