package basket_item

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	servicev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/service/v1"
	mock "go.citilink.cloud/order/internal/specs/grpcclient/mock/citilink/catalog/service/v1"
	"go.citilink.cloud/store_types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestTypeServiceSuite(t *testing.T) {
	suite.Run(t, &TypeServiceSuite{})
}

type TypeServiceSuite struct {
	suite.Suite

	ctrl          *gomock.Controller
	ctx           context.Context
	serviceClient *mock.MockServiceAPIClient
	mapper        *CatalogServicesToDomainMapper
}

func (s *TypeServiceSuite) SetupTest() {}

func (s *TypeServiceSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.serviceClient = mock.NewMockServiceAPIClient(s.ctrl)
	s.mapper = NewCatalogServicesToDomainMapper()
}

func (s *TypeServiceSuite) TestTypeSerive_Determine() {
	type args struct {
		spaceId            store_types.SpaceId
		serviceId          ItemId
		itemGroup          Group
		getToConfiguration bool
	}
	tests := []struct {
		name    string
		args    args
		prepare func(s *TypeServiceSuite)
		want    Type
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "item group invalid",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupInvalid,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {},
			want:    TypeUnknown,
			wantErr: assert.NoError,
		},
		{
			name: "item group product",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupProduct,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {},
			want:    TypeProduct,
			wantErr: assert.NoError,
		},
		{
			name: "item group configuration product",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupProduct,
				getToConfiguration: true,
			},
			prepare: func(s *TypeServiceSuite) {},
			want:    TypeConfigurationProduct,
			wantErr: assert.NoError,
		},
		{
			name: "item group configuration",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupConfiguration,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {},
			want:    TypeConfiguration,
			wantErr: assert.NoError,
		},
		{
			name: "unknown service",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{},
					}, nil).
					Times(1)
			},
			want:    TypeUnknown,
			wantErr: assert.NoError,
		},
		{
			name: "subcontract service for product",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{
							{
								ServiceType: servicev1.ServiceType_SERVICE_TYPE_SUBCONTRACT,
							},
						},
					}, nil).
					Times(1)
			},
			want:    TypeSubcontractServiceForProduct,
			wantErr: assert.NoError,
		},
		{
			name: "insurance service for product",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{
							{
								ServiceType: servicev1.ServiceType_SERVICE_TYPE_INSURANCE_POLICE,
							},
						},
					}, nil).
					Times(1)
			},
			want:    TypeInsuranceServiceForProduct,
			wantErr: assert.NoError,
		},
		{
			name: "property insurance",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{
							{
								ServiceType: servicev1.ServiceType_SERVICE_TYPE_INSURANCE_PRODUCT,
							},
						},
					}, nil).
					Times(1)
			},
			want:    TypePropertyInsurance,
			wantErr: assert.NoError,
		},
		{
			name: "digital service",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{
							{
								Subtype: servicev1.Subtype_SUBTYPE_SOFTWARE_INSTALL,
							},
						},
					}, nil).
					Times(1)
			},
			want:    TypeDigitalService,
			wantErr: assert.NoError,
		},
		{
			name: "configuration assembly service",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{
							{
								Subtype: servicev1.Subtype_SUBTYPE_ASSEMBLY,
							},
						},
					}, nil).
					Times(1)
			},
			want:    TypeConfigurationAssemblyService,
			wantErr: assert.NoError,
		},
		{
			name: "delivery service",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{
							{
								Subtype: servicev1.Subtype_SUBTYPE_DELIVERY,
							},
						},
					}, nil).
					Times(1)
			},
			want:    TypeDeliveryService,
			wantErr: assert.NoError,
		},
		{
			name: "lifting service",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{
							{
								Subtype: servicev1.Subtype_SUBTYPE_RISE,
							},
						},
					}, nil).
					Times(1)
			},
			want:    TypeLiftingService,
			wantErr: assert.NoError,
		},
		{
			name: "not found error",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{},
					}, status.Error(codes.NotFound, "some text")).
					Times(1)
			},
			want:    TypeUnknown,
			wantErr: assert.Error,
		},
		{
			name: "notfound error",
			args: args{
				spaceId:            "msk_cl",
				serviceId:          "someId",
				itemGroup:          GroupService,
				getToConfiguration: false,
			},
			prepare: func(s *TypeServiceSuite) {
				s.serviceClient.EXPECT().
					FindServices(gomock.Any(), gomock.Any()).
					Return(&servicev1.FindServicesResponse{
						Services: []*servicev1.ServiceInfo{},
					}, fmt.Errorf("some text")).
					Times(1)
			},
			want:    TypeUnknown,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			tt.prepare(s)
			serv := NewTypeServices(s.serviceClient, s.mapper)
			got, err := serv.Determine(s.ctx, tt.args.spaceId, tt.args.serviceId, tt.args.itemGroup, tt.args.getToConfiguration)

			if !tt.wantErr(s.T(), err, fmt.Sprintf("Find(%v, %v, %v, %v, %v)", s.ctx, tt.args.spaceId, tt.args.serviceId, tt.args.itemGroup, tt.args.getToConfiguration)) {
				return
			}
			s.Equalf(tt.want, got, "Find(%v, %v, %v, %v, %v)", tt.args.spaceId, tt.args.serviceId, tt.args.itemGroup, tt.args.getToConfiguration)
		})
	}
}
