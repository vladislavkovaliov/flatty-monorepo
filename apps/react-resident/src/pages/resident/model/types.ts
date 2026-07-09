export interface IUser {
  fullName: string;
  email: string;
  phone: string;
  dateOfBirth: string | null;
}

export interface IResident {
  country: string;
  city: string;
  apartment: string;
  house: string;
  street: string;
  postalCode: string;
  address: string;
}

export type ResidentCreate = IUser & IResident;
