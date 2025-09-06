import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { AuthService } from '@services/auth.service';
import { MeService } from '@services/me.service';
import { User } from '@models/user.model';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './navbar.component.html',
  styleUrl: './navbar.component.css'
})
export class NavbarComponent {
  private readonly authService = inject(AuthService);
  private readonly meService = inject(MeService);
  private readonly router = inject(Router);

  user: User | null = null;
  isMenuOpen = false;

  ngOnInit(): void {
    this.loadUserData();
  }

  private loadUserData(): void {
    this.meService.getMeProfile().subscribe({
      next: (user) => {
        this.user = user;
      },
      error: (error) => {
        console.error('Failed to load user data:', error);
        this.user = null; // Reset user data on error
      }
    });
  }

  toggleMenu(): void {
    this.isMenuOpen = !this.isMenuOpen;
  }

  closeMenu(): void {
    this.isMenuOpen = false;
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }

  getInitials(name?: string): string {
    if (!name) return 'U';

    const words = name.trim().split(' ');
    if (words.length === 1) {
      return words[0].charAt(0).toUpperCase();
    } else {
      return (words[0].charAt(0) + words[words.length - 1].charAt(0)).toUpperCase();
    }
  }
}
