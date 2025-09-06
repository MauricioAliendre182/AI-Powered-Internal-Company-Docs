import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';

import { NavbarComponent } from './navbar.component';
import { AuthService } from '@services/auth.service';
import { MeService } from '@services/me.service';
import { provideRouter, Router } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';

describe('NavbarComponent', () => {
  let component: NavbarComponent;
  let fixture: ComponentFixture<NavbarComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockMeService: jasmine.SpyObj<MeService>;
  let router: Router;

  const mockUser = {
    id: '1',
    name: 'John Doe',
    email: 'john@example.com',
    avatar: 'http://example.com/avatar.jpg',
  };

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj('AuthService', ['logout']);
    mockMeService = jasmine.createSpyObj('MeService', ['getMeProfile']);

    // Set up the default behavior
    mockMeService.getMeProfile.and.returnValue(of(mockUser));

    const mockActivatedRoute = {
      snapshot: { params: {}, queryParams: {}, data: {} },
      params: of({}),
      queryParams: of({}),
      data: of({}),
      url: of([]),
      fragment: of(null),
    };

    await TestBed.configureTestingModule({
      imports: [NavbarComponent],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        { provide: MeService, useValue: mockMeService },
        provideRouter([
          // ✅ real router setup so routerLink/routerLinkActive work
          { path: 'login', component: NavbarComponent },
          { path: 'app/documents', component: NavbarComponent },
          { path: 'app/upload', component: NavbarComponent },
          { path: 'app/query', component: NavbarComponent },
          { path: 'app/profile', component: NavbarComponent },
        ]),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
      schemas: [NO_ERRORS_SCHEMA], // ✅ Ignore unknown Angular elements like routerLink
    }).compileComponents();

    fixture = TestBed.createComponent(NavbarComponent);
    component = fixture.componentInstance;

    // Spy on the real router
    router = TestBed.inject(Router);
    spyOn(router, 'navigate');

    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should load user data on init', () => {
    expect(mockMeService.getMeProfile).toHaveBeenCalled();
    expect(component.user).toEqual(mockUser);
  });

  it('should toggle menu', () => {
    expect(component.isMenuOpen).toBeFalse();
    component.toggleMenu();
    expect(component.isMenuOpen).toBeTrue();
    component.toggleMenu();
    expect(component.isMenuOpen).toBeFalse();
  });

  it('should close menu', () => {
    component.isMenuOpen = true;
    component.closeMenu();
    expect(component.isMenuOpen).toBeFalse();
  });

  it('should logout and navigate to login', () => {
    component.logout();
    expect(mockAuthService.logout).toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });

  it('should get correct initials from a full name', () => {
    const initials = component.getInitials('John Doe');
    expect(initials).toBe('JD');
  });

  it('should get correct initial from a single name', () => {
    const initial = component.getInitials('John');
    expect(initial).toBe('J');
  });

  it('should return default initial when name is empty', () => {
    const defaultInitial = component.getInitials();
    expect(defaultInitial).toBe('U');
  });

  it('should handle error when loading user data', () => {
    // Reset the component
    mockMeService.getMeProfile.and.returnValue(
      throwError(() => new Error('API error'))
    );

    spyOn(console, 'error');

    // Trigger ngOnInit again
    component.ngOnInit();

    expect(console.error).toHaveBeenCalledWith(
      'Failed to load user data:',
      jasmine.any(Error)
    );
    expect(component.user).toBeNull();
  });

  it('should display user avatar when available', () => {
    component.user = mockUser;
    fixture.detectChanges();

    const avatar = fixture.debugElement.query(By.css('.user-avatar img'));
    expect(avatar).toBeTruthy();
    expect(avatar.nativeElement.src).toContain(mockUser.avatar);
    expect(avatar.nativeElement.alt).toBe(mockUser.name);
  });

  it('should display user initials when avatar is not available', () => {
    component.user = { ...mockUser, avatar: '' };
    fixture.detectChanges();

    const userAvatar = fixture.debugElement.query(By.css('.user-avatar'));
    expect(userAvatar.nativeElement.textContent.trim()).toBe('JD');
  });
});
