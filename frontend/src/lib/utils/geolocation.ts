export interface GeoResult {
  latitude: number;
  longitude: number;
  accuracy: number;
}

interface GetLocationOptions {
  forceFresh?: boolean;
}

function readPosition(options: PositionOptions): Promise<GeolocationPosition> {
  return new Promise((resolve, reject) => {
    navigator.geolocation.getCurrentPosition(resolve, reject, options);
  });
}

function mapGeolocationError(err: GeolocationPositionError): Error {
  switch (err.code) {
    case err.PERMISSION_DENIED:
      return new Error(
        "Izin lokasi ditolak. Aktifkan lokasi di pengaturan browser.",
      );
    case err.POSITION_UNAVAILABLE:
      return new Error("Lokasi tidak tersedia. Coba ambil lokasi lagi.");
    case err.TIMEOUT:
      return new Error("Timeout mengambil lokasi. Coba ambil lokasi lagi.");
    default:
      return new Error("Gagal mengambil lokasi");
  }
}

function toGeoResult(pos: GeolocationPosition): GeoResult {
  return {
    latitude: pos.coords.latitude,
    longitude: pos.coords.longitude,
    accuracy: pos.coords.accuracy,
  };
}

export async function getLocation(
  options: GetLocationOptions = {},
): Promise<GeoResult> {
  if (!navigator.geolocation) {
    throw new Error("Geolocation tidak didukung di browser ini");
  }

  const primaryOptions: PositionOptions = {
    enableHighAccuracy: true,
    timeout: 15000,
    maximumAge: options.forceFresh ? 0 : 60000,
  };

  try {
    const position = await readPosition(primaryOptions);
    return toGeoResult(position);
  } catch (primaryErr) {
    const err = primaryErr as GeolocationPositionError;
    if (err.code === err.PERMISSION_DENIED) {
      throw mapGeolocationError(err);
    }

    try {
      const fallbackPosition = await readPosition({
        enableHighAccuracy: false,
        timeout: 10000,
        maximumAge: options.forceFresh ? 0 : 300000,
      });
      return toGeoResult(fallbackPosition);
    } catch (fallbackErr) {
      throw mapGeolocationError(fallbackErr as GeolocationPositionError);
    }
  }
}
