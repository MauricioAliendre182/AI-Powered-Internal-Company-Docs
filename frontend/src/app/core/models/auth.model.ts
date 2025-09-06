export interface ResponseLogin {
  message: string;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export interface RegisterData {
  name: string;
  email: string;
  password: string;
}

export type RefreshTokenResponse = Omit<ResponseLogin, 'message'>;
