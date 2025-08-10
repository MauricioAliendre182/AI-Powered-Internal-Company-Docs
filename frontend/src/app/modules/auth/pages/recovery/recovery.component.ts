import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, AbstractControl } from '@angular/forms';
import { Router, RouterModule, ActivatedRoute } from '@angular/router';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import { FaIconLibrary } from '@fortawesome/angular-fontawesome';
import { faExclamationTriangle, faCheckCircle } from '@fortawesome/free-solid-svg-icons';
import { AuthLayoutComponent } from '../../components/auth-layout/auth-layout.component';
import { AuthFormFieldComponent } from '../../components/auth-form-field/auth-form-field.component';
import { AuthButtonComponent } from '../../components/auth-button/auth-button.component';
import { AuthService } from '@services/auth.service';
import { CustomValidators } from '@utils/validators';

@Component({
  selector: 'app-recovery',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterModule,
    FontAwesomeModule,
    AuthLayoutComponent,
    AuthFormFieldComponent,
    AuthButtonComponent
  ],
  templateUrl: './recovery.component.html',
  styleUrl: './recovery.component.css'
})
export class RecoveryComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly library: FaIconLibrary = inject(FaIconLibrary);

  recoveryForm!: FormGroup;
  isLoading = false;
  isCompleted = false;
  token: string | null = null;

  constructor() {
    // Add icons to library
    this.library.addIcons(faExclamationTriangle, faCheckCircle);
    // Subscribe to query params to get the token
    // This is useful if the component is initialized without a token in the URL
    this.route.queryParams.subscribe((params) => {
      const token = params['token'];
      // I need to check if the token exists in the query params
      if (token) {
        // I recover the token from the URL
        this.token = token;
      } else {
        // Redirect to login in case we do not have a token in the URL
        this.router.navigate(['/login']);
      }
    });
  }

  ngOnInit(): void {
    this.initializeForm();
  }

  private initializeForm(): void {
    this.recoveryForm = this.fb.nonNullable.group({
      password: ['', [Validators.required, Validators.minLength(8), CustomValidators.PasswordStrengthValidator()]],
      confirmPassword: ['', [Validators.required]]
    },
    {
      validators: [
        CustomValidators.MatchValidator('password', 'confirmPassword'),
      ],
    },
  );
  }

  onSubmit(): void {
    if (this.recoveryForm.valid && !this.isLoading && this.token) {
      this.isLoading = true;

      const { password } = this.recoveryForm.getRawValue();
      this.authService.changePassword(this.token, password).subscribe({
        next: () => {
          this.isLoading = false;
          this.isCompleted = true;
          this.router.navigate(['/login']);
        },
        error: () => {
          this.isLoading = false;
          this.isCompleted = false;
        },
      });
    } else {
      this.markFormGroupTouched();
    }
  }

  private markFormGroupTouched(): void {
    Object.keys(this.recoveryForm.controls).forEach(key => {
      const control = this.recoveryForm.get(key);
      control?.markAsTouched();
    });
  }

  getFieldError(fieldName: string): string {
    const field = this.recoveryForm.get(fieldName);
    if (field?.errors && field?.touched) {
      if (field.errors['required']) {
        return `${this.getFieldLabel(fieldName)} is required`;
      }
      if (field.errors['minlength']) {
        return `${this.getFieldLabel(fieldName)} must be at least ${field.errors['minlength'].requiredLength} characters`;
      }
      if (field.errors['passwordStrength']) {
        return 'Password must contain uppercase, lowercase, number and special character';
      }
      if (fieldName === 'confirmPassword' && this.recoveryForm.errors?.['passwordMismatch']) {
        return 'Passwords do not match';
      }
    }
    return '';
  }

  private getFieldLabel(fieldName: string): string {
    const labels: { [key: string]: string } = {
      'password': 'Password',
      'confirmPassword': 'Confirm Password'
    };
    return labels[fieldName] || fieldName;
  }

  hasFieldError(fieldName: string): boolean {
    const field = this.recoveryForm.get(fieldName);
    return !!(field?.errors && field?.touched) ||
           (fieldName === 'confirmPassword' && this.recoveryForm.errors?.['passwordMismatch'] && field?.touched);
  }

  onBackToLogin(): void {
    this.router.navigate(['/auth/login']);
  }

  get isTokenValid(): boolean {
    return !!this.token;
  }
}
