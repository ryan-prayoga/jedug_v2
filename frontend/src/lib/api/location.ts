import { apiGet } from './client';
import type { LocationLabelData } from './types';

export async function resolveLocationLabel(latitude: number, longitude: number) {
	const query = new URLSearchParams({
		latitude: String(latitude),
		longitude: String(longitude)
	});

	return apiGet<LocationLabelData>(`/api/v1/location/label?${query.toString()}`);
}
