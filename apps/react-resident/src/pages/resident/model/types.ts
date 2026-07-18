export interface IResident {
  country: string;
  city: string;
  apartment: string;
  house: string;
  street: string;
  postalCode: string;
}

export type ResidentCreate = IResident;
