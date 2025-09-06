import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, AbstractControl } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { AuthLayoutComponent } from '../../components/auth-layout/auth-layout.component';
import { AuthFormFieldComponent } from '../../components/auth-form-field/auth-form-field.component';
import { AuthButtonComponent } from '../../components/auth-button/auth-button.component';
import { AuthService } from '@services/auth.service';
import { RequestStatus } from '@models/request-status.model';
import { CustomValidators } from '@utils/validators';
import { RegisterData } from '@models/auth.model';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterModule,
    AuthLayoutComponent,
    AuthFormFieldComponent,
    AuthButtonComponent
  ],
  templateUrl: './register.component.html',
  styleUrl: './register.component.css'
})
export class RegisterComponent implements OnInit {
  registerForm!: FormGroup;
  isLoading = false;
  status: RequestStatus = 'init';

  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);
  private readonly authService = inject(AuthService);

  ngOnInit(): void {
    this.initializeForm();
  }

  private initializeForm(): void {
    this.registerForm = this.fb.group({
      firstName: ['', [Validators.required, Validators.minLength(2)]],
      lastName: ['', [Validators.required, Validators.minLength(2)]],
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(8), CustomValidators.PasswordStrengthValidator()]],
      confirmPassword: ['', [Validators.required, Validators.minLength(8), CustomValidators.PasswordStrengthValidator()]],
      agreeToTerms: [false, [Validators.requiredTrue]]
    },
    {
      validators: [
        CustomValidators.MatchValidator('password', 'confirmPassword'),
      ],
    },
  );
  }

  onSubmit(): void {
    if (this.registerForm.valid && !this.isLoading) {
      this.isLoading = true;
      const { firstName, lastName, email, password } = this.registerForm.getRawValue();
      const name = `${firstName} ${lastName}`;
      const userData: RegisterData = { name, email, password };
      // I use registerAndLogin to automatically log in the user after registration
      // This improves the user experience by avoiding an extra login step
      // and immediately grants access to the app
      this.authService.registerAndLogin(userData).subscribe({
        next: () => {
          this.status = 'success';
          this.router.navigate(['/app/documents']);
        },
        error: (err) => {
          console.error('Registration failed:', err);
          this.isLoading = false;
          this.status = 'failed';
        },
      });
    } else {
      this.markFormGroupTouched();
    }
  }

  private markFormGroupTouched(): void {
    Object.keys(this.registerForm.controls).forEach(key => {
      const control = this.registerForm.get(key);
      control?.markAsTouched();
    });
  }

  getFieldError(fieldName: string): string {
    const field = this.registerForm.get(fieldName);
    if (field?.errors && field?.touched) {
      if (field.errors['required']) {
        return `${this.getFieldLabel(fieldName)} is required`;
      }
      if (field.errors['email']) {
        return 'Please enter a valid email address';
      }
      if (field.errors['minlength']) {
        return `${this.getFieldLabel(fieldName)} must be at least ${field.errors['minlength'].requiredLength} characters`;
      }
      if (field.errors['passwordStrength']) {
        return 'Password must contain uppercase, lowercase, number and special character';
      }
      if (fieldName === 'confirmPassword' && this.registerForm.errors?.['passwordMismatch']) {
        return 'Passwords do not match';
      }
      if (field.errors['requiredTrue']) {
        return 'You must agree to the terms and conditions';
      }
    }
    return '';
  }

  private getFieldLabel(fieldName: string): string {
    const labels: { [key: string]: string } = {
      'firstName': 'First Name',
      'lastName': 'Last Name',
      'email': 'Email',
      'password': 'Password',
      'confirmPassword': 'Confirm Password'
    };
    return labels[fieldName] || fieldName;
  }

  hasFieldError(fieldName: string): boolean {
    const field = this.registerForm.get(fieldName);
    return !!(field?.errors && field?.touched) ||
           (fieldName === 'confirmPassword' && this.registerForm.errors?.['passwordMismatch'] && field?.touched);
  }
}
