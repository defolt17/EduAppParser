import os
import requests

mshpURL                = "https://my.informatics.ru"
mshpCoursesPageURL     = "/pupil/courses/"
mshpLoginPageURL       = "/accounts/root_login/"
mshpLoginAPIURL        = "/api/v1/rest-auth/login/"
mshpGetClassesAPIURL   = "/api/v1/teaching_situation/classes_users/headings/"
mshpGetClassInfoAPIURL = "api/v1/teaching_situation/classes/extended/"


def getToken():
    pass

userpassword = os.environ['USERPASSWORD']
username = "desyatchenko_if"

response = requests.get(   
    mshpURL,
    headers={'Accept': 'application/json', "Accept-Language": "en-US,en;q=0.5","Origin": "https://my.informatics.ru", "DNT": "1", "__cfduid":	"ddea6b8b8e26a586e730a334bb7d4a7181581584730", "csrftoken": "TStWx9MYuTYwd75KdXywFaBcDSvtFTpsOr2cR2YmZuOdAtVp5aF0dbx64DhBwgVo"},
    #params={'username': username, "password": userpassword, "captcha":  "false",},
)

print(response)