package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2VPCPeeringConnectionAccepter_sameRegionSameAccount(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	resourceNameMainVpc := "aws_vpc.main"                              // Requester
	resourceNamePeerVpc := "aws_vpc.peer"                              // Accepter
	resourceNameConnection := "aws_vpc_peering_connection.main"        // Requester
	resourceNameAccepter := "aws_vpc_peering_connection_accepter.peer" // Accepter
	rName := fmt.Sprintf("terraform-testacc-pcxaccpt-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPCPeeringConnectionAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAccepterSameRegionSameAccountConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceNameAccepter, &connection),
					// The aws_vpc_peering_connection documentation says:
					//	vpc_id - The ID of the requester VPC
					//	peer_vpc_id - The ID of the VPC with which you are creating the VPC Peering Connection (accepter)
					//	peer_owner_id -  The AWS account ID of the owner of the peer VPC (accepter)
					//	peer_region -  The region of the accepter VPC of the VPC Peering Connection
					resource.TestCheckResourceAttrPair(resourceNameConnection, "vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_owner_id", resourceNamePeerVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameConnection, "peer_region", acctest.Region()),
					// The aws_vpc_peering_connection_accepter documentation says:
					//	vpc_id - The ID of the accepter VPC
					//	peer_vpc_id - The ID of the requester VPC
					//	peer_owner_id - The AWS account ID of the owner of the requester VPC
					//	peer_region - The region of the accepter VPC
					// ** TODO
					// ** TODO resourceVPCPeeringRead() is not doing this correctly for same-account peerings
					// ** TODO
					// resource.TestCheckResourceAttrPair(resourceNameAccepter, "vpc_id", resourceNamePeerVpc, "id"),
					// resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_owner_id", resourceNameMainVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameAccepter, "peer_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceNameAccepter, "accept_status", "active"),
				),
			},
			{
				Config:                  testAccVPCPeeringConnectionAccepterSameRegionSameAccountConfig(rName),
				ResourceName:            resourceNameAccepter,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccEC2VPCPeeringConnectionAccepter_differentRegionSameAccount(t *testing.T) {
	var connectionMain, connectionPeer ec2.VpcPeeringConnection
	var providers []*schema.Provider
	resourceNameMainVpc := "aws_vpc.main"                              // Requester
	resourceNamePeerVpc := "aws_vpc.peer"                              // Accepter
	resourceNameConnection := "aws_vpc_peering_connection.main"        // Requester
	resourceNameAccepter := "aws_vpc_peering_connection_accepter.peer" // Accepter
	rName := fmt.Sprintf("terraform-testacc-pcxaccpt-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccVPCPeeringConnectionAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAccepterDifferentRegionSameAccountConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceNameConnection, &connectionMain),
					testAccCheckVPCPeeringConnectionExistsWithProvider(resourceNameAccepter, &connectionPeer, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_owner_id", resourceNamePeerVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameConnection, "peer_region", acctest.AlternateRegion()),
					// ** TODO See TestAccAWSVPCPeeringConnectionAccepter_sameRegion()
					// resource.TestCheckResourceAttrPair(resourceNameAccepter, "vpc_id", resourceNamePeerVpc, "id"),
					// resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_owner_id", resourceNameMainVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameAccepter, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceNameAccepter, "accept_status", "active"),
				),
			},
			{
				Config:                  testAccVPCPeeringConnectionAccepterDifferentRegionSameAccountConfig(rName),
				ResourceName:            resourceNameAccepter,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccEC2VPCPeeringConnectionAccepter_sameRegionDifferentAccount(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	var providers []*schema.Provider
	resourceNameMainVpc := "aws_vpc.main"                              // Requester
	resourceNamePeerVpc := "aws_vpc.peer"                              // Accepter
	resourceNameConnection := "aws_vpc_peering_connection.main"        // Requester
	resourceNameAccepter := "aws_vpc_peering_connection_accepter.peer" // Accepter
	rName := fmt.Sprintf("terraform-testacc-pcxaccpt-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccVPCPeeringConnectionAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAccepterSameRegionDifferentAccountConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceNameConnection, &connection),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_owner_id", resourceNamePeerVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameConnection, "peer_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_owner_id", resourceNameMainVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameAccepter, "peer_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceNameAccepter, "accept_status", "active"),
				),
			},
		},
	})
}

func TestAccEC2VPCPeeringConnectionAccepter_differentRegionDifferentAccount(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	var providers []*schema.Provider
	resourceNameMainVpc := "aws_vpc.main"                              // Requester
	resourceNamePeerVpc := "aws_vpc.peer"                              // Accepter
	resourceNameConnection := "aws_vpc_peering_connection.main"        // Requester
	resourceNameAccepter := "aws_vpc_peering_connection_accepter.peer" // Accepter
	rName := fmt.Sprintf("terraform-testacc-pcxaccpt-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccVPCPeeringConnectionAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAccepterDifferentRegionDifferentAccountConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceNameConnection, &connection),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_owner_id", resourceNamePeerVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameConnection, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_owner_id", resourceNameMainVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameAccepter, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceNameAccepter, "accept_status", "active"),
				),
			},
		},
	})
}

func testAccVPCPeeringConnectionAccepterDestroy(s *terraform.State) error {
	// We don't destroy the underlying VPC Peering Connection.
	return nil
}

func testAccVPCPeeringConnectionAccepterSameRegionSameAccountConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "main" {
  vpc_id      = aws_vpc.main.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCPeeringConnectionAccepterDifferentRegionSameAccountConfig(rName string) string {
	return acctest.ConfigAlternateRegionProvider() + fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "main" {
  vpc_id      = aws_vpc.main.id
  peer_vpc_id = aws_vpc.peer.id
  peer_region = %[2]q
  auto_accept = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.AlternateRegion())
}

func testAccVPCPeeringConnectionAccepterSameRegionDifferentAccountConfig(rName string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "peer" {
  provider = "awsalternate"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "main" {
  vpc_id        = aws_vpc.main.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  peer_region   = %[2]q
  auto_accept   = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.Region())
}

func testAccVPCPeeringConnectionAccepterDifferentRegionDifferentAccountConfig(rName string) string {
	return acctest.ConfigAlternateAccountAlternateRegionProvider() + fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "peer" {
  provider = "awsalternate"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "main" {
  vpc_id        = aws_vpc.main.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  peer_region   = %[2]q
  auto_accept   = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.AlternateRegion())
}
