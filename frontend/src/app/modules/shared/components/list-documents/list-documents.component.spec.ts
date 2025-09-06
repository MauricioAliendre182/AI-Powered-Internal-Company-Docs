import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ListDocumentsComponent } from './list-documents.component';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';

describe('ListDocumentsComponent', () => {
  let component: ListDocumentsComponent;
  let fixture: ComponentFixture<ListDocumentsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ListDocumentsComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ListDocumentsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
