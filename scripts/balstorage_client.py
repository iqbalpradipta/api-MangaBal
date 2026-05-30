import os
from pathlib import Path
from typing import Any

import requests


class BalStorageClient:
    def __init__(self, base_url: str, email: str, password: str, timeout: int = 60):
        self.base_url = base_url.rstrip("/")
        self.email = email
        self.password = password
        self.timeout = timeout
        self.session = requests.Session()

    def login(self) -> None:
        resp = self.session.post(
            f"{self.base_url}/login",
            json={"email": self.email, "password": self.password},
            timeout=self.timeout,
        )
        if resp.status_code >= 400:
            body = resp.text[:500]
            raise RuntimeError(f"BalStorage login failed: HTTP {resp.status_code}: {body}")
        payload = resp.json()
        token = payload.get("data", {}).get("token")
        if not token:
            raise RuntimeError("BalStorage login did not return a token")
        self.session.headers.update({"Authorization": f"Bearer {token}"})

    def list_folders(self, parent_id: str | None = None) -> list[dict[str, Any]]:
        params = {}
        if parent_id:
            params["parent_id"] = parent_id
        resp = self.session.get(f"{self.base_url}/folders", params=params, timeout=self.timeout)
        resp.raise_for_status()
        return self._extract_items(resp.json())

    def ensure_folder(self, name: str, parent_id: str | None = None) -> dict[str, Any]:
        for folder in self.list_folders(parent_id):
            if folder.get("name") == name:
                return folder

        resp = self.session.post(
            f"{self.base_url}/folders",
            json={"name": name, "parent_id": parent_id},
            timeout=self.timeout,
        )
        resp.raise_for_status()
        data = resp.json().get("data")
        if not isinstance(data, dict):
            raise RuntimeError("BalStorage create folder returned unexpected payload")
        return data

    def upload_file(self, folder_id: str, path: str | Path) -> dict[str, Any]:
        file_path = Path(path)
        with file_path.open("rb") as handle:
            files = [("files", (file_path.name, handle, self._mime_for(file_path)))]
            resp = self.session.post(
                f"{self.base_url}/folders/{folder_id}/files",
                files=files,
                timeout=self.timeout,
            )
        resp.raise_for_status()
        items = self._extract_items(resp.json())
        if not items:
            data = resp.json().get("data")
            if isinstance(data, dict):
                return data
            raise RuntimeError("BalStorage upload returned no file data")
        return items[0]

    def file_urls(self, file_id: str) -> dict[str, str]:
        return {
            "preview_url": f"{self.base_url}/files/{file_id}/preview",
            "download_url": f"{self.base_url}/files/{file_id}/download",
            "thumbnail_url": f"{self.base_url}/files/{file_id}/thumbnail",
        }

    @staticmethod
    def _extract_items(payload: dict[str, Any]) -> list[dict[str, Any]]:
        data = payload.get("data")
        if isinstance(data, list):
            return data
        if isinstance(data, dict):
            nested = data.get("data")
            if isinstance(nested, list):
                return nested
            files = data.get("files")
            if isinstance(files, list):
                return files
            folders = data.get("folders")
            if isinstance(folders, list):
                return folders
            if data.get("id"):
                return [data]
        return []

    @staticmethod
    def _mime_for(path: Path) -> str:
        ext = path.suffix.lower()
        if ext in {".jpg", ".jpeg"}:
            return "image/jpeg"
        if ext == ".png":
            return "image/png"
        if ext == ".webp":
            return "image/webp"
        if ext == ".gif":
            return "image/gif"
        return "application/octet-stream"


def file_id_from_upload(upload: dict[str, Any]) -> str:
	for key in ("id", "file_id", "ID"):
		value = upload.get(key)
		if value:
			return str(value)
	file_payload = upload.get("file")
	if isinstance(file_payload, dict):
		for key in ("id", "file_id", "ID"):
			value = file_payload.get(key)
			if value:
				return str(value)
	raise RuntimeError(f"cannot resolve uploaded file id from payload: {upload}")


def file_size_from_upload(upload: dict[str, Any], fallback_path: Path) -> int:
	value = upload.get("size")
	if isinstance(value, int):
		return value
	if isinstance(value, float):
		return int(value)
	file_payload = upload.get("file")
	if isinstance(file_payload, dict):
		value = file_payload.get("size")
		if isinstance(value, int):
			return value
		if isinstance(value, float):
			return int(value)
	return os.path.getsize(fallback_path)


def mime_type_from_upload(upload: dict[str, Any]) -> str:
	value = upload.get("mime_type") or upload.get("mimeType")
	if value:
		return str(value)
	file_payload = upload.get("file")
	if isinstance(file_payload, dict):
		value = file_payload.get("mime_type") or file_payload.get("mimeType")
		if value:
			return str(value)
	return ""
