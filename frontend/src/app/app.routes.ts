import { Routes } from '@angular/router';
import { authGuard } from '@guards/auth.guard';
import { redirectGuard } from '@guards/redirect.guard';

export const routes: Routes = [
  // pathMatch: 'full' is used to redirect the user to the login page if the path is empty
  // and the user is not authenticated
  // pathMatch: 'prefix' is used to redirect the user to the login page if the path is not empty
  // and the user is not authenticated
  {
    path: '',
    redirectTo: 'login',
    pathMatch: 'full',
  },
  {
    path: 'login',
    canActivate: [redirectGuard],
    loadComponent: () => import('./modules/auth/pages/login/login.component').then(m => m.LoginComponent),
    title: 'Login'
  },
  {
    path: 'forgot-password',
    loadComponent: () => import('./modules/auth/pages/forgot-password/forgot-password.component').then(m => m.ForgotPasswordComponent),
    title: 'Forgot your password?'
  },
  {
    path: 'register',
    loadComponent: () => import('./modules/auth/pages/register/register.component').then(m => m.RegisterComponent),
    title: 'Register'
  },
  {
    path: 'recovery',
    loadComponent: () => import('./modules/auth/pages/recovery/recovery.component').then(m => m.RecoveryComponent),
    title: 'Recover Password'
  },
  {
    path: 'app',
    canActivate: [authGuard],
    children: [
      {
        path: '',
        redirectTo: 'documents',
        pathMatch: 'full',
      },
      {
        path: 'documents',
        canActivate: [authGuard],
        loadComponent: () => import('./modules/dashboard/pages/documents/documents.component').then(m => m.DocumentsComponent),
        title: 'Documents'
      },
      {
        path: 'upload',
        canActivate: [authGuard],
        loadComponent: () => import('./modules/dashboard/pages/upload/upload.component').then(m => m.UploadComponent),
        title: 'Upload Document'
      },
      {
        path: 'query',
        canActivate: [authGuard],
        loadComponent: () => import('./modules/dashboard/pages/query/query.component').then(m => m.QueryComponent),
        title: 'Query'
      },
      {
        path: 'profile',
        canActivate: [authGuard],
        loadComponent: () => import('./modules/dashboard/pages/profile/profile.component').then(m => m.ProfileComponent),
        title: 'Profile'
      },
    ],
  },
];
