## 1. GET Area Activity - JSON Response

**Endpoint:** `GET /api/v1/admin/areas/activity?start_date=01-01-2025&end_date=01-01-2025&regional=barat`

**Response:**

```json
{
  "success": true,
  "message": "Area activity retrieved successfully",
  "data": {
    "data": [
      {
        "area_id": 1,
        "area_name": "Segmen I (Cinde-Butik Shinjuku)",
        "regional": "barat",
        "mobil": {
          "masuk": 255,
          "keluar": 248
        },
        "motor": {
          "masuk": 150,
          "keluar": 145
        },
        "total_masuk": 405,
        "total_keluar": 393
      },
      {
        "area_id": 2,
        "area_name": "Parkir Mall Palembang",
        "regional": "barat",
        "mobil": {
          "masuk": 180,
          "keluar": 175
        },
        "motor": {
          "masuk": 90,
          "keluar": 88
        },
        "total_masuk": 270,
        "total_keluar": 263
      }
      // ... area lainnya dengan regional "barat"
    ],
    "summary": {
      "start_date": "2025-01-01",
      "end_date": "2025-01-01",
      "total_areas": 2
    }
  }
}
```

---

## 2. GET Jukir Activity - JSON Response

**Endpoint:** `GET /api/v1/admin/jukirs/activity?start_date=01-01-2025&end_date=01-01-2025&regional=barat`

**Response:**

```json
{
  "success": true,
  "message": "Jukir activity retrieved successfully",
  "data": {
    "data": [
      {
        "jukir_id": 1,
        "jukir_name": "Ahmad",
        "area_id": 1,
        "area_name": "Segmen I (Cinde-Butik Shinjuku)",
        "regional": "barat",
        "mobil": {
          "masuk": 120,
          "keluar": 115
        },
        "motor": {
          "masuk": 80,
          "keluar": 78
        },
        "total_masuk": 200,
        "total_keluar": 193
      },
      {
        "jukir_id": 2,
        "jukir_name": "Budi",
        "area_id": 2,
        "area_name": "Parkir Mall Palembang",
        "regional": "barat",
        "mobil": {
          "masuk": 90,
          "keluar": 88
        },
        "motor": {
          "masuk": 60,
          "keluar": 59
        },
        "total_masuk": 150,
        "total_keluar": 147
      }
      // ... jukir lainnya dengan regional "barat"
    ],
    "summary": {
      "start_date": "2025-01-01",
      "end_date": "2025-01-01",
      "total_jukirs": 2
    }
  }
}
```

---

## 3. Export CSV - Response

**Endpoint:** `GET /api/v1/admin/areas/activity/export?start_date=01-01-2025&end_date=01-01-2025`

**Response:**

```json
{
  "success": true,
  "message": "Area activity exported successfully",
  "data": {
    "filename": "area-activity-2025-01-01-to-2025-01-01.csv",
    "url": "/api/v1/admin/files/exports/activity/1761721525823285708_area-activity-2025-01-01-to-2025-01-01.csv",
    "object_name": "exports/activity/1761721525823285708_area-activity-2025-01-01-to-2025-01-01.csv"
  }
}
```

**Download CSV File:**

```
GET /api/v1/admin/files/exports/activity/1761721525823285708_area-activity-2025-01-01-to-2025-01-01.csv
```

---

## 4. Format CSV File Output

File CSV yang didownload akan memiliki format seperti `docs/contoh-rekap-parkir.csv`:

```
REKAPITULASI DATA PARKIR ;;;;;;;;;;;;
;;;;;;;;;;;;
;;;;;;;;;;;;
Segmen I (Cinde-Butik Shinjuku);;;;Kapasitas ;113;;Segmen I (Cinde-Butik Shinjuku);;;;Kapasitas;34
Mobil Penumpang;Jumlah Kendaraan;;Akumulasi;Volume;Indeks Parkir;;Motor;Jumlah Kendaraan;;Akumulasi;Volume;Indeks Parkir
Periode Pengamatan;Datang;Berangkat;;;;;Periode Pengamatan;Datang;Berangkat;;;
09.00 - 09.15;9;1;8;8;7%;;09.00 - 09.15;1;0;1;1;3%
09.15 - 09.30;21;5;24;32;21%;;09.15 - 09.30;9;1;9;10;26%
09.30 - 09.45;21;13;32;64;28%;;09.30 - 09.45;16;10;15;25;44%
09.45 - 10.00;11;11;32;96;28%;;09.45 - 10.00;11;4;22;47;65%
10.00 - 10.15;4;6;30;126;27%;;10.00 - 10.15;1;5;18;65;53%
...
16.45 - 17.00;6;31;0;891;0%;;16.45 - 17.00;3;21;0;466;0%
;255;;;;;;;150;;;;
```

**Penjelasan Kolom:**

- **Periode Pengamatan**: Interval waktu 15 menit (contoh: "09.00 - 09.15")
- **Datang**: Jumlah kendaraan yang check-in dalam interval tersebut
- **Berangkat**: Jumlah kendaraan yang check-out dalam interval tersebut
- **Akumulasi**: Jumlah kendaraan yang masih parkir di akhir interval
- **Volume**: Total kumulatif check-in sampai akhir interval
- **Indeks Parkir**: Persentase akumulasi terhadap kapasitas (contoh: "7%")

### Catatan:

- Format CSV menggunakan semicolon (`;`) sebagai delimiter
- Setiap area/jukir memiliki section terpisah dalam CSV
- Baris total menampilkan total "Datang" untuk Mobil dan Motor
