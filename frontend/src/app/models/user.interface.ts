export interface User {
  id: string;
  username: string;
  email: string;
  birth_date: string;
  favorite_artist_id?: string;
}

export interface LoginResponse {
  status: string;
  token: string;
  message: string;
}

export interface RegisterResponse {
  status: string;
  token: string;
  message: string;
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  password?: string;
  current_password?: string;
  birth_date?: string;
}

export interface TokenData {
  user_id: string;
  username: string;
  expires: number;
}
