# Frontend

This project was generated using [Angular CLI](https://github.com/angular/angular-cli) version 20.1.0.

## Development server

To start a local development server, run:

```bash
ng serve
```

Once the server is running, open your browser and navigate to `http://localhost:4200/`. The application will automatically reload whenever you modify any of the source files.

## Code scaffolding

Angular CLI includes powerful code scaffolding tools. To generate a new component, run:

```bash
ng generate component component-name
```

For a complete list of available schematics (such as `components`, `directives`, or `pipes`), run:

```bash
ng generate --help
```

## Building

To build the project run:

```bash
ng build
```

This will compile your project and store the build artifacts in the `dist/` directory. By default, the production build optimizes your application for performance and speed.

## Testing Guide

### Running Unit Tests

To execute unit tests with the [Karma](https://karma-runner.github.io) test runner, use the following command:

```bash
ng test
```

To run specific tests, use the `--include` flag:

```bash
ng test --include=**/my-component.spec.ts
```

To run tests without watching for changes (useful in CI environments):

```bash
ng test --no-watch
```

### Testing Best Practices

#### Component Testing

When testing Angular components, follow these best practices:

1. **Test Setup**:
   ```typescript
   describe('MyComponent', () => {
     let component: MyComponent;
     let fixture: ComponentFixture<MyComponent>;
     
     beforeEach(async () => {
       await TestBed.configureTestingModule({
         imports: [MyComponent, /* other dependencies */],
         providers: [/* services */]
       }).compileComponents();
       
       fixture = TestBed.createComponent(MyComponent);
       component = fixture.componentInstance;
       fixture.detectChanges();
     });
     
     it('should create', () => {
       expect(component).toBeTruthy();
     });
   });
   ```

2. **Mocking Services**:
   ```typescript
   const mockService = jasmine.createSpyObj('ServiceName', ['method1', 'method2']);
   mockService.method1.and.returnValue(of(mockData)); // For observables
   ```

3. **HTTP Testing**:
   ```typescript
   // In TestBed setup
   imports: [HttpClientTestingModule],
   providers: [
     provideHttpClient(),
     provideHttpClientTesting()
   ]
   
   // In test
   const httpTestingController = TestBed.inject(HttpTestingController);
   // Make HTTP request
   const req = httpTestingController.expectOne('/api/endpoint');
   req.flush(mockResponseData);
   httpTestingController.verify();
   ```

4. **Router Testing**:
   ```typescript
   // In TestBed setup
   providers: [
     { provide: Router, useValue: jasmine.createSpyObj('Router', ['navigate']) },
     provideRouter([]) // For components with RouterModule imports
   ]
   ```

5. **Testing DOM Interactions**:
   ```typescript
   const button = fixture.debugElement.query(By.css('.my-button'));
   button.nativeElement.click();
   fixture.detectChanges();
   expect(component.wasClicked).toBeTrue();
   ```

6. **Testing Async Code**:
   ```typescript
   it('should load data asynchronously', fakeAsync(() => {
     component.loadData();
     tick(); // Wait for async operations
     expect(component.data).toEqual(mockData);
   }));
   ```

7. **Testing Component with Router Issues**:
   When testing components with router directives (routerLink, routerLinkActive), 
   you may need to use one of these approaches:
   
   - Use `NO_ERRORS_SCHEMA` to ignore unknown attributes:
     ```typescript
     TestBed.configureTestingModule({
       schemas: [NO_ERRORS_SCHEMA]
     });
     ```
     
   - Test component class methods in isolation:
     ```typescript
     it('should test method in isolation', () => {
       const testComponent = new MyComponent();
       const mockDep = jasmine.createSpyObj('Dependency', ['method']);
       (testComponent as any).dependency = mockDep;
       
       testComponent.methodToTest();
       expect(mockDep.method).toHaveBeenCalled();
     });
     ```

## Running end-to-end tests

For end-to-end (e2e) testing, run:

```bash
ng e2e
```

Angular CLI does not come with an end-to-end testing framework by default. You can choose one that suits your needs.

### Troubleshooting Common Testing Issues

1. **Router Issues**: If you encounter `Cannot read properties of undefined (reading 'root')` errors when testing components with router dependencies, try:
   - Using `NO_ERRORS_SCHEMA`
   - Testing component class methods directly
   - Providing a mock Router: `{ provide: Router, useValue: mockRouter }`
   - Using `provideRouter([])` alongside your mock Router

2. **HttpClient Issues**: If you see `No provider for HttpClient` errors:
   - Add `HttpClientTestingModule` to your imports
   - Use `provideHttpClient()` and `provideHttpClientTesting()`

3. **Async Testing Issues**: For testing code with Observables:
   - Use `fakeAsync` and `tick()` to control async timing
   - Use `waitForAsync` for asynchronous test completion

## Additional Resources

For more information on using the Angular CLI, including detailed command references, visit the [Angular CLI Overview and Command Reference](https://angular.dev/tools/cli) page.

For Angular testing documentation, visit:
- [Angular Testing Guide](https://angular.dev/guide/testing)
- [Testing Components](https://angular.dev/guide/testing/components)
