
$ErrorActionPreference = 'Stop';
$packageName = 'octant'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64 = 'https://github.com/restuhaqza/octant/releases/download/v0.26.0/octant_0.26.0_Windows-64bit.zip'
$checksum64 = 'b1fd33d49616c9bd7ed78e20813ca60b03afbb7f779051e45a182271587ccd6d'
$checksumType64= 'sha256'

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url64bit      = $url64

  softwareName  = 'octant*'

  checksum64    = $checksum64
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs

