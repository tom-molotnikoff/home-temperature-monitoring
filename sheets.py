from apiclient import discovery
from google.oauth2 import service_account
import json

def convert_reading_to_sheets_value(reading):
    return [list(reading.values())]


class Sheets:
    sheet_id = ""
    service = 0

    def __init__(self, sheet_id, secret_file_path):
        self.sheet_id = sheet_id
        try:
            scopes = ["https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/drive.file",
                      "https://www.googleapis.com/auth/spreadsheets"]
            credentials = service_account.Credentials.from_service_account_file(secret_file_path, scopes=scopes)
            self.service = discovery.build('sheets', 'v4', credentials=credentials)
        except OSError as e:
            print(e)

    def insert_data(self, reading, range_name):
        """
        Take a reading from a sensor and input into a Google Sheet at a specified range
        """
        body_val = convert_reading_to_sheets_value(reading)
        print(body_val)
        self.service.spreadsheets().values().update(spreadsheetId=self.sheet_id, body=body_val, range=range_name,
                                                    valueInputOption='USER_ENTERED').execute()
