import { Component, input, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-auth-layout',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './auth-layout.component.html',
  styleUrl: './auth-layout.component.css'
})
export class AuthLayoutComponent {
  title = input<string>(''); // Using input() for standalone component
  subtitle = input<string>('');
  showBackToLogin = input<boolean>(false);
  backToLoginText = input<string>('Back to Login');
  backToLoginRoute = input<string>('/login');
}
