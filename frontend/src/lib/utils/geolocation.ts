export interface GeoResult {
  latitude: number;
  longitude: number;
  accuracy: number;
}

export function getLocation(): Promise<GeoResult> {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error("Geolocation tidak didukung di browser ini"));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        resolve({
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
          accuracy: pos.coords.accuracy,
        });
      },
      (err) => {
        switch (err.code) {
          case err.PERMISSION_DENIED:
            reject(
              new Error(
                "Izin lokasi ditolak. Aktifkan lokasi di pengaturan browser.",
              ),
            );
            break;
          case err.POSITION_UNAVAILABLE:
            reject(new Error("Lokasi tidak tersedia. Coba lagi."));
            break;
          case err.TIMEOUT:
            reject(new Error("Timeout mengambil lokasi. Coba lagi."));
            break;
          default:
            reject(new Error("Gagal mengambil lokasi"));
        }
      },
      {
        enableHighAccuracy: true,
        timeout: 15000,
        maximumAge: 60000,
      },
    );
  });
}
