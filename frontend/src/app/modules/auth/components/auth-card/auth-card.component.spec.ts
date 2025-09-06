import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { Component, ViewChild } from '@angular/core';
import { AuthCardComponent } from './auth-card.component';

// Create a test host component to properly provide inputs
@Component({
  template: `
    <app-auth-card
      [title]="title"
      [subtitle]="subtitle"
      [showHeader]="showHeader"
      [showFooter]="showFooter">
      <div class="content-test">Test Content</div>
    </app-auth-card>
  `,
  standalone: true,
  imports: [AuthCardComponent]
})
class TestHostComponent {
  @ViewChild(AuthCardComponent) authCardComponent!: AuthCardComponent;
  title = 'Test Title';
  subtitle = 'Test Subtitle';
  showHeader = true;
  showFooter = false;
}

describe('AuthCardComponent', () => {
  let hostComponent: TestHostComponent;
  let fixture: ComponentFixture<TestHostComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TestHostComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TestHostComponent);
    hostComponent = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(hostComponent.authCardComponent).toBeTruthy();
  });

  it('should display title and subtitle when provided', () => {
    const titleElement = fixture.debugElement.query(By.css('.card-title'));
    const subtitleElement = fixture.debugElement.query(By.css('.card-subtitle'));

    expect(titleElement.nativeElement.textContent).toContain('Test Title');
    expect(subtitleElement.nativeElement.textContent).toContain('Test Subtitle');
  });

  it('should not display header when showHeader is false', () => {
    hostComponent.showHeader = false;
    fixture.detectChanges();

    const headerElement = fixture.debugElement.query(By.css('.card-header'));
    expect(headerElement).toBeNull();
  });

  it('should not display title when title is empty', () => {
    hostComponent.title = '';
    fixture.detectChanges();

    const titleElement = fixture.debugElement.query(By.css('.card-title'));
    expect(titleElement).toBeNull();
  });

  it('should not display subtitle when subtitle is empty', () => {
    hostComponent.subtitle = '';
    fixture.detectChanges();

    const subtitleElement = fixture.debugElement.query(By.css('.card-subtitle'));
    expect(subtitleElement).toBeNull();
  });

  it('should display footer when showFooter is true', () => {
    hostComponent.showFooter = true;
    fixture.detectChanges();

    const footerElement = fixture.debugElement.query(By.css('.card-footer'));
    expect(footerElement).toBeTruthy();
  });

  it('should not display footer when showFooter is false', () => {
    const footerElement = fixture.debugElement.query(By.css('.card-footer'));
    expect(footerElement).toBeNull();
  });

  it('should always display card body', () => {
    const bodyElement = fixture.debugElement.query(By.css('.card-body'));
    expect(bodyElement).toBeTruthy();
  });

  it('should project content into the card body', () => {
    const contentElement = fixture.debugElement.query(By.css('.content-test'));
    expect(contentElement).toBeTruthy();
    expect(contentElement.nativeElement.textContent).toContain('Test Content');
  });

  it('should handle all input combinations', () => {
    // Test with only title
    hostComponent.title = 'Only Title';
    hostComponent.subtitle = '';
    fixture.detectChanges();

    let titleElement = fixture.debugElement.query(By.css('.card-title'));
    let subtitleElement = fixture.debugElement.query(By.css('.card-subtitle'));

    expect(titleElement).toBeTruthy();
    expect(subtitleElement).toBeNull();

    // Test with only subtitle
    hostComponent.title = '';
    hostComponent.subtitle = 'Only Subtitle';
    fixture.detectChanges();

    titleElement = fixture.debugElement.query(By.css('.card-title'));
    subtitleElement = fixture.debugElement.query(By.css('.card-subtitle'));

    expect(titleElement).toBeNull();
    expect(subtitleElement).toBeTruthy();

    // Test with both empty but header showing
    hostComponent.title = '';
    hostComponent.subtitle = '';
    hostComponent.showHeader = true;
    fixture.detectChanges();

    const headerElement = fixture.debugElement.query(By.css('.card-header'));
    expect(headerElement).toBeNull(); // Header should not render if both title and subtitle are empty
  });
});
