import { TestBed } from '@angular/core/testing';

import { MeService } from './me.service';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';

describe('MeService', () => {
  let service: MeService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      // providers is to make sure the service and HttpClientTestingModule are available
      // provideHttpClient is needed for Angular 15+
      // provideHttpClientTesting is needed for HttpTestingController
      providers: [
        MeService,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    service = TestBed.inject(MeService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
