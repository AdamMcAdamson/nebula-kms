---

This is a general outline of the endpoints we need to implement for the KMS service.
These endpoints are not final.


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
    NewName
}
Return:
{
    Name,
    Last_Modified
}


PATCH /Regenerate-Key
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID
}
Return:
{
    Key,
    Last_Modified
}


POST /Create-Basic-Key
FROM DeveloperPortalBackend
{
    UserID
}
Return:
{
    Key_Mongo_OID,
    Name,
    Key,
    Timed_Quota,
    Usage_Remaining,
    Key_Created
}


POST /Create-Advanced-Key
FROM DeveloperPortalBackend
{
    CreatorUserID,
    RecipientUserID,
    Service_Mongo_OID,
    Quota,
    Quota_Interval_Type
}
Return:
{
    Key_Mongo_OID,
    Name,
    Key_Created
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
    Key_Mongo_OID
}
Return:
{
    "Success" | "Error"
}


PATCH /Enable-Key
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID
}
Return:
{
    "Success" | "Error"
}


PATCH /Set-Key-Quota
FROM DeveloperPortalBackend
{
    UserID,
    Key_Mongo_OID,
    Quota,
    Quota_Interval_Type
}
Return:
{
    Usage_Remaining,
    Quota_Timestamp,
    Last_Modified
}


*
GET /logs
FROM DeveloperPortalBackend
{
    UserID
}
Return:
{

}


---