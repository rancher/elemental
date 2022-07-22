export class TopLevelMenu {
  toggle() {
    cy.get('[data-testid="top-level-menu"]', {timeout: 12000}).click();
  }

  openIfClosed() {
    cy.get('body').then((body) => {
      if (body.find('.menu.raised').length === 0) {
        this.toggle();
      }
    });
  }

  categories() {
    cy.get('.side-menu .body .category');
  }

  links() {
    cy.get('.side-menu .option');
  }

  clusters(clusterName: string) {
    cy.get('.clusters .cluster.selector.option').contains(clusterName).click();
  }

  localization() {
    cy.get('.locale-chooser');
  }
}
