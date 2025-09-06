import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, provideHttpClientTesting } from '@angular/common/http/testing';

import { UploadDocumentComponent} from './upload-document.component';
import { provideHttpClient } from '@angular/common/http';

describe('UploadDocumentComponent', () => {
  let component: UploadDocumentComponent;
  let fixture: ComponentFixture<UploadDocumentComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [
        UploadDocumentComponent
      ],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    })
    .compileComponents();

    fixture = TestBed.createComponent(UploadDocumentComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
