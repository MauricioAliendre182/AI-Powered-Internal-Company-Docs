import { ApplicationConfig, provideBrowserGlobalErrorListeners, provideZoneChangeDetection } from '@angular/core';
import { provideRouter } from '@angular/router';

import { routes } from './app.routes';
import { authInterceptor } from '@interceptors/auth.interceptor';
import { provideHttpClient, withFetch, withInterceptors } from '@angular/common/http';

export const appConfig: ApplicationConfig = {
  providers: [
    // provicdeBrowserGlobalErrorListeners() is used to provide global error listeners for the application.
    // This is useful for handling errors that occur in the application.
    provideBrowserGlobalErrorListeners(),
    // provideZoneChangeDetection() is used to configure change detection in the application.
    // It can be used to optimize change detection performance.
    provideZoneChangeDetection({ eventCoalescing: true }),
    // provideRouter() is used to configure the router for the application.
    // It takes the routes defined in app.routes.ts and sets up the router.
    provideRouter(routes),
    // Importing the token interceptor to be used in the application.
    // This interceptor will handle token management for HTTP requests.
    // Here we are applying my interceptor to all HTTP requests.
    // This is a good place to add interceptors that should be applied to all requests.
    provideHttpClient(withInterceptors([authInterceptor])),
    // This line is to provide the HTTP client with fetch capabilities.
    // It allows the application to use the fetch API for making HTTP requests.
    provideHttpClient(withFetch())
  ]
};
