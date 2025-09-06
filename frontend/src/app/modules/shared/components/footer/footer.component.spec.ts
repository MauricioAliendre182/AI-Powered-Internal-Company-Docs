import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';

import { FooterComponent } from './footer.component';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

describe('FooterComponent', () => {
  let component: FooterComponent;
  let fixture: ComponentFixture<FooterComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FooterComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    }).compileComponents();

    fixture = TestBed.createComponent(FooterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should have the current year', () => {
    const currentYear = new Date().getFullYear();
    expect(component.currentYear).toEqual(currentYear);
  });

  it('should display the footer with company info', () => {
    const footerElement = fixture.debugElement.query(By.css('.footer'));
    expect(footerElement).toBeTruthy();

    const brandText = fixture.debugElement.query(By.css('.brand-text'));
    expect(brandText.nativeElement.textContent).toContain('DocAI');
  });

  it('should have a footer-description', () => {
    const description = fixture.debugElement.query(By.css('.footer-description'));
    expect(description).toBeTruthy();
    expect(description.nativeElement.textContent.trim()).toContain('AI-powered');
  });

  it('should have quick links section', () => {
    const quickLinksSection = fixture.debugElement.query(By.css('.footer-title'));
    expect(quickLinksSection).toBeTruthy();
    expect(quickLinksSection.nativeElement.textContent).toContain('Quick Links');
  });
});
