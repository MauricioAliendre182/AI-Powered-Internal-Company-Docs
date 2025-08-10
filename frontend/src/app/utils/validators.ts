import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';

export class CustomValidators {
  // This MatchValidator checks if the value of one form control matches the value of another.
  // It is useful for validating fields like password confirmation.
  static MatchValidator(source: string, target: string): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      const sourceCtrl = control.get(source);
      const targetCtrl = control.get(target);

      return sourceCtrl && targetCtrl && sourceCtrl.value !== targetCtrl.value
        ? { mismatch: true }
        : null;
    };
  }

  // This PasswordStrengthValidator checks if the password meets certain strength criteria.
  // It checks for the presence of uppercase letters, lowercase letters, numbers, and special characters
  // and returns an error if the password does not meet these criteria.
  // It is useful for ensuring that user passwords are strong enough.
  static PasswordStrengthValidator(): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      const value = control.value;
      if (!value) return null;

      const hasUpperCase = /[A-Z]/.test(value);
      const hasLowerCase = /[a-z]/.test(value);
      const hasNumeric = /\d/.test(value);
      const hasSpecialChar = /[!@#$%^&*(),.?":{}|<>]/.test(value);

      const valid = hasUpperCase && hasLowerCase && hasNumeric && hasSpecialChar;
      return valid ? null : { passwordStrength: true };
    };
  }
}
