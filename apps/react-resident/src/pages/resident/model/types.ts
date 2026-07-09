export interface IResidentCreate {
  fullName: string;
  email: string;
  phone: string;
  dateOfBirth: string;
  country: string;
  city: string;
  address: string;
}

export interface IResidentLocation {
  id: number;
  country: string;
  city: string;
}

export interface IResidentLocationResponse {
  data: IResidentLocation[];
  total: number;
}