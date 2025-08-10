import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DashboardLayoutComponent } from '@shared/components/dashboard-layout/dashboard-layout.component';
import { MeService } from '@services/me.service';
import { User } from '@models/user.model';
import { RequestStatus } from '@models/request-status.model';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, DashboardLayoutComponent],
  templateUrl: './profile.component.html',
  styleUrl: './profile.component.css'
})
export class ProfileComponent implements OnInit {
  private readonly meService = inject(MeService);

  user: User | null = null;
  profileStatus: RequestStatus = 'init';
  error = '';

  ngOnInit(): void {
    this.loadProfile();
  }

  loadProfile(): void {
    this.profileStatus = 'loading';
    this.error = '';

    this.meService.getMeProfile().subscribe({
      next: (user) => {
        this.user = user;
        this.profileStatus = 'success';
      },
      error: (error) => {
        this.profileStatus = 'failed';
        this.error = error.error?.error || 'Failed to load profile';
        console.error('Failed to load profile:', error);
      }
    });
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

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }
}
