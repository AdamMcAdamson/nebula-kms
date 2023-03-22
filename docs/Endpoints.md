```js

// This is a general outline of the endpoints we need to implement for the KMS service.
// These endpoints are not final, but their functionality is well defined.

GET /allowed
FROM gateway
{
    Source-Identifier,
    Key
}
Return:
{
    "Success" | "Error"
}


GET /KMS-User-Keys
FROM DeveloperPortalBackend
{
    UserID
}
Return:
{
    User_Collection * Key_Collection
}


GET /KMS-Privileged-Data
FROM DeveloperPortalBackend
{
    UserID
}
Return:
{
    User_Collection * Service (Filtered by User Type) * Key_Collection (Limited)
}


GET /User-Type
FROM DeveloperPortalBackend
{
    UserID
}
Return:
{
    User_Type
}


PATCH /Rename-Key
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID,
    NewName,
    Last_Modified
}
Return:
{
    Name,
    Last_Modified
}, {409}// old Last_Modified


PATCH /Regenerate-Key
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID,
    Last_Modified
}
Return:
{
    Key,
    Last_Modified
}, {409}// old Last_Modified


POST /Create-Basic-Key
FROM DeveloperPortalBackend
{
    UserID
}
Return:
{
    Key_Mongo_OID,
    Key,
    Timed_Quota,
    Usage_Remaining,
    Key_Created,
    Last_Modified
}


POST /Create-Advanced-Key
FROM DeveloperPortalBackend
{
    CreatorUserID,
    RecipientUserID,
    Service_Mongo_OID,
    Quota,
    Quota_Interval_Type,
    Name
}
Return:
{
    Key_Mongo_OID,
    Name,
    Key_Created,
    Last_Modified
}


DELETE /Delete-Key
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID
}
Return:
{
    "Success" | "Error"
}


PATCH /Disable-Key
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID,
    Last_Modified
}
Return:
{
    "Success" | "Error",
    Last_Modified
}, {409}// old Last_Modified


PATCH /Enable-Key
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID,
    Last_Modified
}
Return:
{
    "Success" | "Error",
    Last_Modified
}, {409}// old Last_Modified


PATCH /Set-Key-Quota
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID,
    Quota,
    Quota_Interval_Type,
    Last_Modified
}
Return:
{
    Usage_Remaining,
    Quota_Timestamp,
    Last_Modified
}, {409}// old Last_Modified

PATCH /Restore-Key-Quota
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID,
    Last_Modified
}
Return:
{
    Usage_Remaining,
    Quota_Timestamp,
    Last_Modified
}, {409}// old Last_Modified

PATCH /Key-Change-Service
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID,
    Service_Mongo_OID,
    Last_Modified
}
Return:
{
    Service_Mongo_OID,
    Last_Modified
}, {409}// old Last_Modified


PATCH /Change-Key-Holder
FROM DeveloperPortalBackend
{
    AssignerUserID,
    RecipientUserID,
    Key_Mongo_OID,
    Last_Modified,
    Name
}
Return:
{
    Key_Mongo_OID,
    Name,
    Last_Modified
}, {409}// old Last_Modified


// *TBD
GET /logs
FROM DeveloperPortalBackend
{
    UserID
}
Return:
{

}


```
