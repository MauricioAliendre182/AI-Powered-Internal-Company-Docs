import { Component, forwardRef, input, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ControlValueAccessor, NG_VALUE_ACCESSOR, ReactiveFormsModule } from '@angular/forms';

@Component({
  selector: 'app-auth-form-field',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './auth-form-field.component.html',
  styleUrl: './auth-form-field.component.css',
  providers: [
    // Provide ControlValueAccessor for this component
    // This allows the component to be used as a form control
    // and integrate with Angular forms
    // NG_VALUE_ACCESSOR is an Angular token that allows us to register this component as a form control
    // The forwardRef is used to avoid circular dependency issues
    // multi: true allows multiple providers for the same token
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => AuthFormFieldComponent),
      multi: true
    }
  ]
})
export class AuthFormFieldComponent implements ControlValueAccessor {
  label = input<string>(''); // Using input() for standalone component
  type = input<string>('text');
  placeholder = input<string>('');
  icon = input<string>('');
  errorMessage = input<string>('');
  showPassword = input<boolean>(false);
  hasError = input<boolean>(false);
  required = input<boolean>(false);
  disabled = input<boolean>(false);
  disabledButton = signal<boolean>(false); // Using signal for reactive state management

  value: string = '';
  showPasswordText: boolean = false;
  isFocused: boolean = false;
  fieldId: string = `field-${Math.random().toString(36).substring(2, 11)}`;

  // ControlValueAccessor methods
  onChange = (value: string) => {};
  onTouched = () => {};

  writeValue(value: string): void {
    this.value = value || '';
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    if (this.disabled() == false) {
      this.disabledButton.set(isDisabled);
    }

  }

  onInput(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.value = target.value;
    this.onChange(this.value);
  }

  onFocus(): void {
    this.isFocused = true;
  }

  onBlur(): void {
    this.isFocused = false;
    this.onTouched();
  }

  togglePasswordVisibility(): void {
    this.showPasswordText = !this.showPasswordText;
  }

  get inputType(): string {
    if (this.type() === 'password') {
      return this.showPasswordText ? 'text' : 'password';
    }
    return this.type();
  }
}
